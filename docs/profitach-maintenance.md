# ProfiTach maintenance baseline

Initial upstream baseline: `way-platform/tachograph-go` **v0.18.2**,
`95ca68093ad6e7d25deb7c0de0c45a3c9abcedab`. Established and validated on
2026-11-10. See the [upstream PR audit](upstream-pr-audit.md) for the reviewed
patches and ordered recommendations.

## Local patch

`internal/dd/full_card_number.go` adds CONTROL_CARD to the existing owner branch
in decoding, encoding and string formatting. This is equivalent to the three
production hunks in upstream #41 and the CONTROL_CARD subset of #19. Neither PR
was merged wholesale.

The binary representation is unchanged: 18 bytes comprising equipment type
`0x03`, one issuing-state byte, 13 owner-identification bytes, and one byte each
for the consecutive, replacement and renewal indices. The protobuf equipment
value is **5**, explaining `unsupported card type: 5`; it must not be written
as byte 5 on the wire. Data Dictionary sections 2.26 and 2.73 specify the owner
choice for control bodies. Driver cards retain their separate 14+1+1 layout.

The regression tests use synthetic identifiers and buffers:

- All four supported card types decode and round-trip with and without retained
  raw data; owner/driver choice and each index are checked independently.
- CONTROL_CARD encoding from constructed semantic fields works with no raw data
  and with a blank raw canvas, without mutating that canvas.
- Owner formatting, null/space padding, invalid lengths, absent/unrecognized
  card handling and recognized non-card rejection are covered.
- A synthetic complete Gen1 overview with a control activity parses and
  round-trips. Gen2 V1 and V2 control RecordArrays parse and re-encode their
  32-byte records with both Gen1 and Gen2 card-generation bytes.

Tests were first run against the original implementation, including a Go source
overlay of the v0.18.2 file after the tests were finalized. They reproduced the
error in Gen1 control activity and both Gen2 paths; independent encoding also
showed the missing owner bytes. The same tests pass with the three-switch fix.

This does not implement full control-card file support, change authentication,
or fix the pre-existing direct cast of the issuing-state byte. Nation mapping is
separate work identified in #19. Existing absent-card normalization and formatting
length behavior are unchanged. Synthetic signatures in the overview test are
placeholders for parsing, not cryptographic verification fixtures.

## Compatibility and data handling

- Preserve `github.com/way-platform/tachograph-go` and all nested module paths.
  The original MIT license and Way Data Technologies copyright remain unchanged.
- Keep each future patch isolated, with its upstream PR/head and rationale
  recorded. Re-check whether a patch is already upstream before replaying it.
  CONTROL_CARD must not be applied again from #19, #41 or #32.
- Authenticate original RawFile bytes before semantic parsing. Do not preprocess
  or rewrite original downloads to work around parser errors. Write derived
  artifacts separately; original signatures do not authenticate anonymized data.
- Never commit real DDD/V1B downloads, their full-file goldens, or raw PII extracts.
  Use synthetic regressions or deterministically anonymized record fixtures.
  Review nested GNSS/time data as well as visible names, plates and card numbers.
- Require generation-specific binary and semantic tests for parsing changes.
  Raw passthrough alone cannot demonstrate correct semantic reconstruction.
  Changes to protobuf field meanings or tags require a migration review.

## Validation results

All checks below passed on Linux amd64. The root and tools modules used Go
1.25.1; CLI modules used Go 1.25.0, following their existing module declarations.
Go was available at `/home/victor/.local/go/bin` but initially absent from PATH.
No module or dependency declarations were changed. golangci-lint v2.10.0 was
installed outside the repository (its build selected Go 1.26.10).

| Check | Scope | Result |
| --- | --- | --- |
| Focused regressions against original v0.18.2 source overlay | `internal/dd`, `internal/vu` | Expected failures reproduced |
| `go test -json ./...` | Root module | Pass; existing fixture skips below |
| `go test -race -json ./...` | Root module | Pass; no races reported |
| `go test -race -json ./internal/dd ./internal/vu` | Final tests, including independent encoding added after the full run | Pass |
| `go test -race ./...` | Each of `cli`, `cmd/tachograph`, `tools` | Pass; these packages have no test files |
| `go vet ./...` | Each of all four modules | Pass |
| `go build ./...` | Each of all four modules, including CLI executable | Pass |
| `golangci-lint run --timeout=5m` | Each of all four modules | Pass, zero issues; no automatic fixes |
| `gofmt`, `git diff --check` | Changed Go files / patch | Pass |

Equivalent reproducible commands, with Go and golangci-lint on PATH:

```sh
go test ./...
go test -race ./...
for module in . cli cmd/tachograph tools; do
  (cd "$module" && go vet ./... && go build ./... && golangci-lint run --timeout=5m)
done
for module in cli cmd/tachograph tools; do
  (cd "$module" && go test -race ./...)
done
git diff --check
```

`mise` was not installed, so checks ran directly. The combined `mise run build`
wrapper was not executed: proto formatting/codegen and module tidy were not
needed for this Go-only production change, and its final diff gate would reject
intentional uncommitted changes. Generated code, dependencies, module paths,
license and existing fixture files are unchanged. No upstream branch other than
the baseline plus our patch was executed as part of validation.

## Fixture gaps and skipped tests

No proprietary full-file downloads were available in this checkout. These
integration tests skipped under both normal and race runs:

- Root: `TestUnmarshalFile_golden`.
- Card: `TestInferCardFileType`, `TestDriverCardFileGolden`,
  `Test_roundTrip_rawCardFile`,
  `TestUnmarshalOptions_UnmarshalRawCardFile_golden`.
- VU: `TestRawVehicleUnitFileGolden`, `TestRawVehicleUnitFileRoundTrip`,
  `TestUnmarshalVehicleUnitFile`, `TestVehicleUnitFileGolden`.

The committed fixture corpus contains no Gen2 V1 overview, activities,
events/faults or technical-data hexdumps. Consequently `TestOverview_Gen2V1`,
`TestActivities_Gen2V1`, `TestEventsAndFaults_Gen2V1` and
`TestTechnicalData_Gen2V1` skipped. The new synthetic Gen2 V1 CONTROL_CARD test
runs independently of those fixtures; it does not replace complete VU coverage.

The Gen2 ERCA certificate
`internal/cert/certcache/g2/18250066869740371713.bin` is absent. The following
ECC tests skipped: `TestVerifyEccCertificateWithCA/Finland_MSCA_Card42`,
`TestVerifyEccCertificateWithCA_MultipleCountries` and
`TestVerifyEccCertificateWithCA_InvalidSignature`. Existing available security
tests passed, but this is not full real-chain verification coverage. No trust
material was fetched or changed to bypass the skips.

## Completion and follow-up

Completed: isolated CONTROL_CARD fix, focused regressions, all-open-PR audit,
maintenance guidance, unit/race/vet/build/lint validation and skip inventory.

Pending follow-up: private real-world CONTROL_CARD verification with authorized
fixtures, missing Gen2 V1 record coverage and ERCA fixture provisioning. The
remaining upstream integrations are recommendations only; their compatibility,
privacy and cryptographic gates are documented in the audit. No commits, merges,
release tags or remote changes were made in this task.
