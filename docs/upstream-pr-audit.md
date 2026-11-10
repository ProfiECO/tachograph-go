# ProfiTach upstream PR audit

Reviewed 2026-11-10 against upstream **v0.18.2**, commit
`95ca68093ad6e7d25deb7c0de0c45a3c9abcedab`. The working tree started clean at
that tag. GitHub's open-PR query returned **11 PRs**, all listed below, including
#44 beyond the requested list. Recommendations concern the fetched heads, not
just their titles or original descriptions. No unrelated patch was integrated.

Evidence: upstream PR metadata and per-file patches, compared with local baseline
implementations and tests. The linked PRs are the primary sources; reviewed head
SHAs below make the snapshot explicit. Upstream test code was reviewed, but the
unintegrated branches and their claimed private download corpora were **not run**.
Our validation applies to v0.18.2 plus the CONTROL_CARD fix; see
[maintenance and validation](profitach-maintenance.md).

## Decision table

“Present” refers to the original v0.18.2 baseline, before our fix. Existing helper
functions or schemas do not mean that a proposed behavior is already integrated.

| PR and purpose | Affected packages / functionality | Present in baseline? | Compatibility / correctness risks | Tests and missing evidence | Recommendation | Dependencies / conflicts |
| --- | --- | --- | --- | --- | --- | --- |
| [#41](https://github.com/way-platform/tachograph-go/pull/41): CONTROL_CARD owner identification | `internal/dd`; all callers of FullCardNumber, including VU control activities | No | Low: protocol byte 3 must select the 13+1+1+1 owner layout; proto enum value is 5. Must cover formatting and encoding, not merely suppress the error. | Upstream synthetic decode/round-trip test; claimed three private VDO downloads unavailable. Local regressions additionally cover each index, formatting, independent encoding, raw painting, padding, other card types, and Gen1/Gen2 VU paths. | **Integrate — completed**, equivalent three-switch change with expanded synthetic tests. | Same CONTROL_CARD hunks as #19; apply once. Supersedes #32's partial-card workaround. |
| [#19](https://github.com/way-platform/tachograph-go/pull/19): VU compatibility, nation mappings, calibration alignment and cyclic limit | `internal/dd`, `internal/vu`, `internal/card`, VU proto/generated code and goldens | No proposed fixes; protocol mapping already exists in `UnmarshalVehicleRegistration`, but not the other nation paths changed here | Much broader than title: 14-byte registration normalization changes output size; skipped arrays/extensions lack semantic preservation; calibration field types change at existing proto tags; buffer-derived activity limit changes termination behavior. | Nation and calibration tests and golden updates exist. Private 2,185-file success claim is not reproducible here. Missing complete Gen2 V1 corpus, unknown-array/truncation tests, extension-preserving semantic write-back and proto migration evidence. | **Selectively cherry-pick** isolated, proven fixes; CONTROL_CARD portion already done. Defer broad structural/schema changes. | Overlaps #41 exactly in three switches; #32 offers contradictory 14-byte interpretation; #23 touches VU sizing; #26 assumes fixed cyclic limit; #42 edits some of the same golden files. |
| [#22](https://github.com/way-platform/tachograph-go/pull/22): certificate validity dates | `internal/security`; Gen1 RSA and Gen2 ECC verification, new clock abstraction | No; dates are parsed but not checked against wall time | Default `time.Now()` changes historical-download authentication as certificates age. Clock injection is internal and not exposed through public AuthenticateOptions. Skipping nil dates requires distinguishing legitimate sentinels from missing required metadata. | Twelve test functions cover clocks, dates and certificate paths; ECC tests can skip without ERCA fixtures. Missing public API historical-verification policy, exact-boundary/invalid-timestamp and full chain/cache behavior evidence. | **Defer** until verification-time policy is explicit and tested. | Coordinate with #25 in `ecc_verify.go`; neither substitutes for signature or trust-chain validation. |
| [#23](https://github.com/way-platform/tachograph-go/pull/23): RecordArray overflow / bounds protection | New `internal/safemath`; VU raw sizing and Gen1/Gen2 parser files | No | Needed hardening, but generic multiplication misses `(-1, MinInt)` overflow; header/offset addition also needs checks. Some checks occur after record loops. Zero-size records with nonzero count can still reach fixed slices. | Arithmetic and nine VU tests; two expected products exceed 32-bit `int` at compile time. Missing a passing 386 run, negative/large offsets, addition boundaries, zero-size/nonzero-count parser tests and fuzz evidence. | **Selectively cherry-pick**, fixing arithmetic and tests first; high priority. | Apply before #19's dynamic array scanning; reconcile VU files with #29. |
| [#24](https://github.com/way-platform/tachograph-go/pull/24): activity minute/count validation | `internal/dd/activity_change_info.go`, `internal/card/activity.go`; both card generations, shared VU decoder | No | Rejects formerly accepted malformed data. Count excludes skipped sentinel entries; no lower bound or trailing-byte validation. Existing card parser skips `0x0000`, although this also encodes a valid midnight rest state. Encoder remains unchecked. | Five minute boundary tests and three count tests. Missing sentinel/midnight preservation, zero changes, odd-length records, marshal symmetry and malformed Gen1/Gen2 VU tests. | **Selectively cherry-pick** with explicit strictness and sentinel handling. | Add #26 coverage first; shares `activity.go` with #19 but edits different logic. |
| [#25](https://github.com/way-platform/tachograph-go/pull/25): ECC CAR/CHR consistency | `internal/security/ecc_verify.go`; Gen2 certificate-chain identity | No; analogous Gen1 RSA check exists | Useful identity constraint; silently skips empty CAR/CHR. A matching reference is not proof of authenticity: ECDSA, trusted roots and key checks remain necessary. | Three test functions exercise mismatch, empty values and CA delegation with stub keys. Missing successful real/synthetic signed chain, tampered signature with matching CAR, absent required references and supported-curve coverage. | **Selectively cherry-pick** after stronger verification tests and explicit missing-reference policy. | Same verifier as #22; independent of card parsing. |
| [#26](https://github.com/way-platform/tachograph-go/pull/26): cyclic iterator tests | `internal/card/cyclic_buffer_test.go`; backward ring traversal | No direct test file; iterator already has cycle detection, oldest-position stop and 5,000-record limit | Tests preserve existing behavior, including silent cycle/header termination; they do not establish that behavior is spec-correct. | Eleven tests: wrap, nil, lengths, cycle, backstop, oldest position. Missing split headers across ring end, record longer than buffer, oversized previous length, exactly-limit termination and semantic chronology/reconstruction tests. | **Integrate** as a separate test-only change after adapting to project comparison style. Not applied in this task. | Fixed 5,000 expectation conflicts with #19's buffer-derived limit. Establish coverage before changing cyclic behavior. |
| [#29](https://github.com/way-platform/tachograph-go/pull/29): Gen2 EF codecs/wiring and sentinel normalization | `internal/card`, `internal/dd`, VU activities; places/GNSS authentication, border crossings, load/unload, load type, company activity | Partial foundations only: schemas, DD helpers and border-crossing codec exist; proposed dispatch/unparse/anonymization and sentinel changes do not | Nested GNSS data survives load/unload anonymization; odometer getter is set unconditionally there. Nil sentinels can normalize bytes and presence. New EFs are emitted with nil signatures. Company-card EF placement in driver dispatch needs validation. | Synthetic EF/parser tests and optional fixture discovery, plus coordinate/odometer tests. Missing full signed parse→unparse→marshal comparison, recursive privacy tests, ring-pointer validation, mixed-generation dispatch, and one-coordinate-only sentinel fidelity. | **Selectively cherry-pick** EF-by-EF after privacy and round-trip fixes; split global sentinel changes. | Needs correct FullCardNumber handling (#41); conflicts with #23 in VU activity files, and benefits from #19 nation fixes. |
| [#32](https://github.com/way-platform/tachograph-go/pull/32): Crono permissive workarounds | `internal/dd`, `internal/vu/technical_data_gen2_v1.go`, `PATCHES.md` | No | Returns control cards without owner ID; skips non-168-byte calibration arrays, including unvalidated declared spans. Treats 14-byte registration as nation+13 characters, conflicting with #19's codepage+13 characters. Drops semantic data. | No added tests. Private firmware observations do not resolve layout ambiguity or establish safe round-trip behavior. | **Reject as submitted**. Retain the reports as investigation leads. | Superseded by #41 for control cards; conflicts with #19 registration/calibration decoding and #23 bounds hardening. |
| [#42](https://github.com/way-platform/tachograph-go/pull/42): Gen1 event/fault enum mappings | `internal/dd/vu_fault_record.go` shared helpers; Gen1 VU event/fault goldens | No; baseline directly casts bytes | Corrects semantic labels without changing known protocol bytes. Invalid enum marshaling still silently falls back to zero; unrecognized values use the existing UNSPECIFIED+raw-value convention. Downstream consumers may rely on old incorrect JSON. | Known-value and unknown event-type round trips; three Gen1 golden updates. Missing unknown-purpose and invalid semantic marshal tests; exhaustively checking all 256 byte values would strengthen confidence. | **Integrate** next as an isolated correctness fix with semantic release note and extra edge tests. Not applied here. | Independent of #41; golden-file overlap with #19 requires regeneration from anonymized hexdumps, not textual merging. |
| [#44](https://github.com/way-platform/tachograph-go/pull/44): dependency refresh | Root and `tools` module files; CEL/protovalidate transitives, `x/net` in certificate-fetch tooling | No; baseline CEL 0.26.1 and `x/net` 0.41.0 | CEL evaluation/validation behavior and toolchain/transitive requirements may change. No binary parsing fix. | No added tests. Need all four modules' test/lint/build checks against upgraded dependencies, Go 1.25 compatibility and vulnerability review. | **Defer** to a separate dependency maintenance change. | No direct code conflict; test #29's presence changes with CEL validation if combined. |

## Integration order and gates

1. **Done:** minimal #41-equivalent CONTROL_CARD fix. Do not replay #19's same
   switches or #32's fallback over it.
2. **Next:** #42 semantic enum correction; verify all protocol values and unknown
   purposes, regenerate Gen1 goldens and notify consumers of corrected JSON.
3. **Hardening:** repair/select #23; require amd64 and 386 tests before introducing
   more dynamic RecordArray parsing. Check sizes against remaining bytes *before*
   allocation or slicing, including offsets and headers.
4. **Trust chain:** select #25 with positive signed-chain and negative tampering
   tests. Separate identity matching from cryptographic signature verification.
5. **Cyclic safety:** #26 tests, then selected #24 checks. Resolve midnight/sentinel
   semantics and header wrapping before any #19 cyclic-limit change.
6. **Gen2 completeness:** split #29 into small EF patches, with recursive
   anonymization and signed binary round-trip gates. Handle sentinel normalization
   separately, including absence versus zero and preservation of original bytes.
7. **Broader VU compatibility:** split remaining #19 nation mapping, registration,
   and calibration work. Require anonymized generation-specific evidence and a
   schema migration plan. Do not import its test identifiers described as real
   vehicle data; replace them with deterministic synthetic identifiers.
8. **Separate maintenance:** evaluate #44 dependencies/toolchain requirements.
   Revisit #22 only after agreeing on historical versus current-time verification
   policy. #32 remains rejected as submitted.

These are recommendations, not authorization to merge the remaining PRs.

## Specific compatibility findings

- **#19 versus #41:** both fetched heads add CONTROL_CARD to decoding, encoding
  and string formatting. #41's description emphasizes two paths, but its patch
  changes all three. #19 additionally changes issuing-state interpretation; that
  is a separate baseline defect, deliberately left unchanged by our minimal fix.
- **#19 schema:** calibration fields 5 and 6 change from
  `FullCardNumberAndGeneration` to `FullCardNumber`, and `Date` to `Timestamp`.
  Reusing those field numbers with incompatible embedded-message meanings is not
  a safe protobuf migration. Generated Go accessors and JSON names also change.
  Correcting binary offsets should be separated from schema evolution.
- **#19 structural fidelity:** accepting a 222-byte calibration record while
  marshaling a 168-byte base loses extensions without raw passthrough. A successful
  raw round trip does not prove semantic editing or anonymization is lossless.
  Unknown-array skips need bounds validation, generation-aware ordering and a
  preservation strategy. The changed offsets end at byte 167 while the minimum
  and output remain 168; reconcile sizes with the regulation before adoption.
- **#29 privacy:** `anonymizeLoadUnloadOperations` copies
  `GnssPlaceAuthRecord` directly, retaining coordinates and its nested timestamp.
  It also copies odometer values regardless of `PreserveDistanceAndTrips` and
  sets absent odometers to zero. The new anonymization code needs recursive tests
  before being used to produce committed fixtures.
- **Certificates:** no trust store or signature algorithm was changed here.
  Do not infer authentication success from successful semantic parsing. Keep
  verification over original RawFile bytes before parsing; modifications or
  anonymization cannot retain the original signature's validity.

## Reviewed heads

| PR | Head SHA |
| --- | --- |
| #19 | `ec7f6120fc60d4da74ea220dccd3aeb8b8c8c806` |
| #22 | `33d477822be7c54574a87d0b7ae1294a5c069af3` |
| #23 | `ec99e7f35b809900e2b3f4bdc4e1840978ba7767` |
| #24 | `b3e0da7b9b1e136e0ab1a98cdf62326378ecc86c` |
| #25 | `aa7d53f8d2d1eba4e70a6de5daa0136e300771c7` |
| #26 | `b54c0d2bb16cc9bad57e83eba0f5c8f352a70650` |
| #29 | `863889e8197e9eb29f51b581b7ffb34b935284ba` |
| #32 | `ebeca85b7fe00064ab4f9a2740c98d1b85b264c5` |
| #41 | `5d26972eb3c7183ab625c39d4fee9cb3ba8c80a6` |
| #42 | `3f32c5611580b69d3799e5fbb9404ec5e7f84d60` |
| #44 | `17d635d928b7384ec6a22836dd917c0fb94763f1` |
