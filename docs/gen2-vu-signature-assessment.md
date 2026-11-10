# Gen2 VU signature verification: issue #39 assessment

Assessed 2026-11-10 against upstream v0.18.2 and the maintained fork after
`4867ed4`. [Upstream issue #39](https://github.com/way-platform/tachograph-go/issues/39)
remains open, with no comments at review time. Its private 20-file corpus was not
available. **Assessment complete; authentication repair remains pending.**
No authentication policy, trust material, signature algorithm or success status
was relaxed in this maintenance queue.

## Confirmed defects

| Area | Baseline and current behavior | Evidence / consequence |
| --- | --- | --- |
| Signature RecordArray wrapper | `internal/vu/authenticate.go`, `verifyGen2DataSignature`, passes the entire signature slice to `security.VerifyEccDataSignature`. Gen2 sizers include the five-byte RecordArray header in `signature_size`. | A 64-byte P-256 signature becomes 69 bytes (`08 0040 0001` + signature) and is rejected with `got 69, want 64`. Still present. |
| Overview signed range | The same function hashes all data preceding the signature. It has no Gen2 Overview exception for the two certificate RecordArrays. | The local regulation reference, Appendix 7 §2.2.6.2, excludes certificates from Overview signed data. `extractGen2Certificates` already computes the two variable-length array spans for extraction, but signature verification does not reuse those boundaries. Certificate extraction itself is not the reported signed-range defect. Still present in both Gen2 Overview paths. |
| TechnicalData boundaries | `sizeOfTechnicalDataGen2V2` counts seven arrays and labels the seventh as the signature without inspecting its type. The semantic parser and marshal path also model seven arrays, omitting `VuCardRecordArray`. | In an eight-array sequence, the seventh array is power-supply interruption (`0x1f`), and the eighth is the real signature (`0x08`). A synthetic reproduction reports 788 signature bytes and leaves the actual 69-byte signature outside the transfer value. Still present. |
| Related Gen2 V1 boundary risk | `sizeOfTechnicalDataGen2V1` counts four arrays; the matching semantic parser uses the same reduced structure. | The local amended TREP 25 reference lists eight arrays, including external GNSS, VU cards, ITS consent and power interruptions. This needs a generation/amendment-aware review, not an assumption that every V1 download has one fixed count. |

The #23 bounds changes prove that declared arrays fit in memory/input. They do
**not** prove that array types, ordering, semantic layouts or signature boundaries
are correct. Positional parsing still accepts a wrong array as a signature.

The local text extraction of Appendix 7 contains the V1 array table but embeds the
V2 table as an image whose asset is absent. Before implementing complete V2
framing, check the authoritative consolidated regulation's TREP 35 table directly;
do not derive its entire schema solely from the issue or local positional parser.

## Diagnostics performed

A temporary Go test was run through a source overlay, without adding a test that
permanently requires broken behavior to the regression suite. It generated a
synthetic NIST P-256 key and signed a non-personal test payload with Go's ECDSA
signer. No key or download was saved in the repository.

- Bare `r || s` verified successfully through `security.VerifyEccDataSignature`.
- Altered payload, altered signature and a 63-byte signature were rejected.
- Passing the valid signature in `08 0040 0001` framing through
  `verifyGen2DataSignature` reproduced the 69-versus-64 error and
  `DATA_SIGNATURE_INVALID` status.
- To isolate the second defect, the diagnostic supplied a bare signature with
  `signature_size=64`, preceded by two synthetic certificate array spans and the
  signed payload. The payload alone verified; including the prefixes failed.
  These prefixes were not certificates and no certificate chain was claimed.
- A framing-only sample contained empty headers of types
  `19,20,21,0c,0e,17`, then `1f 0057 0009` plus 783 bytes, then a 69-byte
  signature array. The current V2 sizer consumed **818 of 887 bytes**, reporting
  the **788-byte `0x1f` array** as the signature. The payloads are deliberately
  structural test data, not semantically valid TechnicalData records.

These results establish the local failure mechanisms, not interoperability or
successful end-to-end VU authentication.

## Available fixtures and blockers

- Two public Finnish MSCA ECC certificates exist under
  `internal/security/testdata/certs/g2`: `finland_msca_card42.bin` and
  `finland_msca_card43.bin`. Certificate parse/round-trip tests can run. The cache
  has `1316820541130145537.bin`, an MSCA certificate, not the required ERCA root.
- ERCA fixture `internal/cert/certcache/g2/18250066869740371713.bin` is absent.
  Existing positive chain, multiple-country and tampered-signature chain tests
  skip where they require it. Available CAR-mismatch coverage is not a positive
  trusted-chain fixture.
- The two committed Gen2 V2 TechnicalData hexdumps (records directories 003 and
  004) each contain 1,144 bytes, seven sequential array tags `01..07`, and a final
  zero-size/zero-count signature placeholder. They exercise current parsing and
  raw round trips; they are not cryptographic positives or independent evidence
  of the actual protocol RecordType assignments.
- No complete original signed Gen2 VU downloads or verified matching VU/MSCA/ERCA
  chain are available in this checkout. Existing full-file integration tests skip.
  Anonymized downloads cannot be substituted: changing signed data invalidates
  original signatures.
- Generated P-256 positives and tampering negatives are available as a diagnostic
  technique. They can become permanent isolated regression fixtures in the repair,
  but do not replace independently verified V1/V2 device data, Brainpool coverage,
  or a complete test certificate chain.

**Validation blocker:** a complete repair cannot be claimed until there are
independently checked generation-specific signed byte ranges and positive VU
signatures with a trusted chain. Obtain authorized private originals locally
without committing them, or create a fully synthetic signed VU and test CA chain
checked against the regulation and an independent verifier. Never upload real
PII as a workaround, and never classify an anonymized original signature as valid.

## Isolated implementation stages

1. **Establish fixtures and byte-range oracles.** Confirm TREP 21/31/25/35 framing
   against the consolidated regulation. Document exact signed bytes, certificate
   array spans, signature array offset and expected result for each fixture.
   Generate test-only keys and certificate chains; keep public trust roots distinct
   from test trust configuration. Include standard P-256 and the supported
   Brainpool curves used by real files.
2. **Strict signature extraction.** Require RecordType `0x08`, count exactly one,
   curve-appropriate record size, complete payload and no trailing bytes. Reuse the
   checked array-span parser. Reject zero-count anonymized placeholders. Do not
   adopt the issue's permissive helper: it accepts arbitrary types/multiple
   signatures and falls back to treating malformed wrappers as bare signatures.
3. **Overview signed-range selection.** For both Gen2 generations, validate and
   skip exactly the MSCA and VU certificate arrays, including their headers, while
   keeping them available for chain verification. Use their encoded spans, not
   re-marshaled certificate lengths. Hash the original remaining bytes. Do not
   apply the certificate offset to activities, events/faults or speed transfers.
4. **TechnicalData framing and semantic preservation.** Implement the confirmed
   ordered/type-aware array layout, including VU card records, separately from
   signature extraction. Reconcile V1 amendments and V2 layouts. Raw sizing,
   semantic parsing, unmarshaling/marshaling and anonymization must agree. Do not
   scan arbitrary bytes for `0x08`, silently drop unknown arrays, or skip data to
   make a signature validate. Treat proto evolution as separate compatibility work.
5. **End-to-end authentication gate.** Verify actual original RawFile bytes through
   the public authentication API, including certificate resolution and chain checks.
   Positive signatures must verify; every negative must remain non-VERIFIED.
   Only then announce Gen2 VU authentication support as repaired.

Required regressions include: wrapper sizes/counts/types, truncated headers and
payloads, missing/duplicate/non-final signature arrays, offset overflow, wrong
certificate order/count, wrong key, changed data/signature, changed signed array
headers, unsupported curves, missing trust roots, wrong CAR/CHR, missing VU-card
arrays and real boundaries across adjacent TV records. Assert unchanged raw bytes,
correct error statuses, Gen1/driver-card compatibility, and both Gen2 versions.
Repeat race, amd64/386 bounds tests and all module suites after each stage.

Until those gates pass, the next maintenance tag must explicitly retain the
Gen2 VU authentication limitation. Parser improvements are not a security fix.
