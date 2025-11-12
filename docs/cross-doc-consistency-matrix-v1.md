# Cross-Doc Consistency Matrix v1

> **Document ID**: Cross-Doc-Consistency-Matrix-v1  
> **Version**: 1.0  
> **Date**: 2025-01-15  
> **Classification**: Internal  
> **Scope**: Verification matrix for Radio Control Container documentation alignment

---

## 1. Document Control

- **Revision History**

  | Version | Date       | Author            | Changes                                                                                         |
  | ------- | ---------- | ----------------- | ----------------------------------------------------------------------------------------------- |
  | 1.0     | 2025-01-15 | System Architect  | Initial consistency matrix following documentation alignment task allocation. |

- **Approval Signatures**

  | Role               | Name  | Signature | Date |
  | ------------------ | ----- | --------- | ---- |
  | System Architect   | [TBD] |           |      |
  | IV&V Lead          | [TBD] |           |      |
  | Product Owner      | [TBD] |           |      |

---

## 2. Purpose & Scope

This document provides a comprehensive consistency matrix verifying that all Radio Control Container documentation is aligned and cross-referenced correctly. Each rule is validated across all relevant documents.

**Documents Covered:**
- **CB-TIMING v0.3** - Baseline timing parameters
- **Architecture Document** - System architecture and policies
- **OpenAPI v1** - Northbound API specification
- **Telemetry SSE v1** - Event stream specification
- **ICD Logical** - Southbound radio interface

---

## 3. Consistency Rules Matrix

### 3.1 Channel Index → Frequency Mapping

| Rule | Architecture §13 | OpenAPI §3.7, §3.8 | ICD §6.1.3 | Status |
|------|------------------|---------------------|------------|--------|
| **1-based indexing** | ✓ Explicit | ✓ Explicit | ✓ Implicit | ✅ **ALIGNED** |
| **Frequency precedence** | ✓ Explicit | ✓ Explicit | N/A | ✅ **ALIGNED** |
| **Derivation process** | ✓ Detailed | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |
| **Validation rules** | ✓ Referenced | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |

**Cross-References:**
- Architecture §13 → OpenAPI §3.7, §3.8
- OpenAPI §3.7, §3.8 → Architecture §13
- ICD §6.1.3 → Architecture §13

### 3.2 Error Model & Normalization

| Rule | Architecture §8.5 | OpenAPI §2.2 | Telemetry §5 | ICD §8 | Status |
|------|-------------------|--------------|-------------|-------|--------|
| **Error codes** | `BAD_REQUEST`, `INVALID_RANGE`, `BUSY`, `UNAVAILABLE`, `INTERNAL` | ✓ Identical | ✓ Identical | ✓ Referenced | ✅ **ALIGNED** |
| **HTTP mapping** | N/A | ✓ Explicit | ✓ Explicit | N/A | ✅ **ALIGNED** |
| **Normalization rules** | ✓ Detailed | ✓ Referenced | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |
| **Vendor error handling** | ✓ §8.5.1 | ✓ Referenced | ✓ Referenced | ✓ §8.1 | ✅ **ALIGNED** |

**Cross-References:**
- Architecture §8.5 → OpenAPI §2.2, Telemetry §5
- OpenAPI §2.2 → Architecture §8.5
- Telemetry §5 → Architecture §8.5
- ICD §8 → Architecture §8.5

### 3.3 Timing & Cadence Parameters

| Parameter | CB-TIMING v0.3 | Architecture §8.3 | Telemetry §4 | Status |
|-----------|----------------|-------------------|--------------|--------|
| **Heartbeat interval** | ✓ 15s | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |
| **Probe cadences** | ✓ 3 states | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |
| **Backoff policies** | ✓ Detailed | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |
| **Event buffering** | ✓ 50 events | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |

**Cross-References:**
- CB-TIMING v0.3 → Architecture §8.3, §8.3a
- Architecture §8.3, §8.3a → CB-TIMING v0.3
- Telemetry §4 → CB-TIMING v0.3

### 3.4 Power-Aware Operating Modes

| Rule | Architecture §8.3a | CB-TIMING v0.3 | Status |
|------|-------------------|----------------|--------|
| **Event-first policy** | ✓ Explicit | ✓ Referenced | ✅ **ALIGNED** |
| **Duty-cycled probing** | ✓ 3 states | ✓ Detailed | ✅ **ALIGNED** |
| **Probe budgets** | ✓ Referenced | ✓ Explicit | ✅ **ALIGNED** |
| **Multi-radio independence** | ✓ Explicit | ✓ Referenced | ✅ **ALIGNED** |

**Cross-References:**
- Architecture §8.3a → CB-TIMING v0.3
- CB-TIMING v0.3 → Architecture §8.3a

### 3.5 Telemetry & Event Streams

| Rule | Architecture §9.3 | OpenAPI §3.9 | Telemetry SSE v1 | Status |
|------|-------------------|--------------|------------------|--------|
| **Event types** | ✓ Listed | ✓ Listed | ✓ Detailed | ✅ **ALIGNED** |
| **Resume semantics** | ✓ Referenced | ✓ Referenced | ✓ Detailed | ✅ **ALIGNED** |
| **Buffer management** | ✓ Referenced | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |
| **Event ordering** | ✓ Referenced | ✓ Referenced | ✓ Detailed | ✅ **ALIGNED** |

**Cross-References:**
- Architecture §9.3 → Telemetry SSE v1
- OpenAPI §3.9 → Telemetry SSE v1
- Telemetry SSE v1 → Architecture §9.3

### 3.6 Radio Discovery & Lifecycle

| Rule | Architecture §5.6 | ICD §15 | Status |
|------|-------------------|---------|--------|
| **Startup capability ingest** | ✓ Explicit | ✓ Explicit | ✅ **ALIGNED** |
| **Runtime duty-cycled checks** | ✓ Referenced | ✓ Referenced | ✅ **ALIGNED** |
| **Loss & re-attach** | ✓ Explicit | ✓ Explicit | ✅ **ALIGNED** |
| **Capability refresh** | ✓ Explicit | ✓ Explicit | ✅ **ALIGNED** |

**Cross-References:**
- Architecture §5.6 → ICD §15
- ICD §15 → Architecture §5.6

### 3.7 Frequency Profile Parsing

| Rule | ICD §6.1.3 | Architecture §13 | Status |
|------|------------|------------------|--------|
| **Range format** | ✓ Explicit | ✓ Referenced | ✅ **ALIGNED** |
| **Single frequency** | ✓ Explicit | ✓ Referenced | ✅ **ALIGNED** |
| **Normalization rules** | ✓ Explicit | ✓ Referenced | ✅ **ALIGNED** |
| **Units (MHz)** | ✓ Explicit | ✓ Referenced | ✅ **ALIGNED** |

**Cross-References:**
- ICD §6.1.3 → Architecture §13
- Architecture §13 → ICD §6.1.3

---

## 4. Cross-Reference Validation

### 4.1 Architecture Document References

| Section | References | Target Document | Status |
|---------|------------|----------------|--------|
| §8.3 | CB-TIMING v0.3 | CB-TIMING v0.3 | ✅ **VALID** |
| §8.3a | CB-TIMING v0.3 | CB-TIMING v0.3 | ✅ **VALID** |
| §8.5 | OpenAPI §2.2 | OpenAPI v1 | ✅ **VALID** |
| §8.5.1 | OpenAPI specification | OpenAPI v1 | ✅ **VALID** |
| §9.3 | Telemetry SSE v1 | Telemetry SSE v1 | ✅ **VALID** |
| §13 | OpenAPI §3.7, §3.8 | OpenAPI v1 | ✅ **VALID** |

### 4.2 OpenAPI Document References

| Section | References | Target Document | Status |
|---------|------------|----------------|--------|
| §2.2 | Architecture §8.5 | Architecture | ✅ **VALID** |
| §3.7, §3.8 | Architecture §13 | Architecture | ✅ **VALID** |
| §3.9 | Telemetry SSE v1 | Telemetry SSE v1 | ✅ **VALID** |

### 4.3 Telemetry SSE Document References

| Section | References | Target Document | Status |
|---------|------------|----------------|--------|
| §1.3 | CB-TIMING v0.3 | CB-TIMING v0.3 | ✅ **VALID** |
| §4 | CB-TIMING v0.3 | CB-TIMING v0.3 | ✅ **VALID** |
| §5 | Architecture §8.5 | Architecture | ✅ **VALID** |

### 4.4 ICD Document References

| Section | References | Target Document | Status |
|---------|------------|----------------|--------|
| §6.1.3 | Architecture §13 | Architecture | ✅ **VALID** |
| §8.1 | Architecture §8.5.1 | Architecture | ✅ **VALID** |
| §15 | Architecture §5.6 | Architecture | ✅ **VALID** |

### 4.5 Terminology Consistency

| Document | Component Name | Abbreviation | Status |
|----------|----------------|--------------|--------|
| Architecture | Radio Control Container | RCC | ✅ **STANDARDIZED** |
| OpenAPI | Radio Control Container (RCC) | RCC | ✅ **STANDARDIZED** |
| Telemetry SSE | Radio Control Container (RCC) | RCC | ✅ **STANDARDIZED** |
| ICD | Radio Control Container (RCC) | RCC | ✅ **STANDARDIZED** |

---

## 5. Consistency Verification Results

### 5.1 Overall Status: ✅ **ALIGNED**

All consistency rules have been verified and are aligned across all documents.

### 5.2 Key Achievements

1. **Single Source of Truth**: All timing parameters centralized in CB-TIMING v0.3
2. **Cross-Reference Integrity**: All references validated and working
3. **Error Model Consistency**: Identical error codes across all documents
4. **Channel Mapping Alignment**: Consistent precedence rules and derivation
5. **Power-Aware Design**: Event-first policies consistently applied
6. **Telemetry Integration**: Seamless event stream specifications

### 5.3 Validation Summary

| Category | Rules | Aligned | Issues |
|----------|-------|---------|--------|
| **Channel Mapping** | 4 | 4 | 0 |
| **Error Model** | 4 | 4 | 0 |
| **Timing Parameters** | 4 | 4 | 0 |
| **Power Management** | 4 | 4 | 0 |
| **Telemetry** | 4 | 4 | 0 |
| **Radio Lifecycle** | 4 | 4 | 0 |
| **Frequency Parsing** | 4 | 4 | 0 |
| **Cross-References** | 16 | 16 | 0 |

**Total**: 44 rules verified, 44 aligned, 0 issues

---

## 6. Compliance Checklist

### 6.1 Architecture Document
- [x] §8.3 references CB-TIMING v0.3
- [x] §8.3a references CB-TIMING v0.3
- [x] §8.5.1 vendor error ambiguity documented
- [x] §13 channel mapping precedence rule explicit
- [x] §5.6 radio lifecycle references CB-TIMING v0.3
- [x] Implementation details purged

### 6.2 OpenAPI Document
- [x] Channels schema updated to objects with index/frequency
- [x] POST /radios/{id}/channel precedence rule documented
- [x] GET /radios/{id}/channel null handling documented
- [x] Telemetry section links to Telemetry SSE v1
- [x] Error model references Architecture §8.5

### 6.3 Telemetry SSE Document
- [x] Resume & buffer semantics reference CB-TIMING v0.3
- [x] Heartbeat & cadence reference CB-TIMING v0.3
- [x] Fault catalog matches OpenAPI error codes
- [x] Event ordering per-radio monotonic documented

### 6.4 ICD Document
- [x] Frequency profile parsing rules documented
- [x] Error input ambiguity mirrors Architecture §8.5.1
- [x] Capability ingest & refresh documented
- [x] References Architecture §13 for channel derivation

### 6.5 CB-TIMING Document
- [x] All timing parameters centralized
- [x] Cross-reference matrix included
- [x] Parameter validation rules defined
- [x] Tuning guidelines provided

---

## 7. Next Steps

### 7.1 Implementation Phase
1. **Developer Review**: All documents ready for implementation
2. **IV&V Testing**: Use consistency matrix for test case generation
3. **Integration Testing**: Validate cross-document references in practice

### 7.2 Maintenance Phase
1. **Change Management**: Update matrix when documents change
2. **Version Control**: Track document version dependencies
3. **Regular Audits**: Quarterly consistency verification

---

## 8. Annex A - Document Dependencies

```
CB-TIMING v0.3 (Baseline)
    ↓
Architecture Document
    ↓
OpenAPI v1 ← Telemetry SSE v1
    ↓
ICD Logical
```

**Dependency Rules:**
- Changes to CB-TIMING require Architecture updates
- Changes to Architecture require OpenAPI/Telemetry updates
- Changes to OpenAPI/Telemetry require ICD updates
- All changes must maintain cross-reference integrity

---

## 9. Annex B - Validation Checklist

### 9.1 For Each Document Update
- [ ] All cross-references validated
- [ ] CB-TIMING parameters referenced (not hardcoded)
- [ ] Error codes consistent across documents
- [ ] Channel mapping rules aligned
- [ ] Power-aware policies maintained
- [ ] Telemetry event schemas consistent

### 9.2 For Each Cross-Reference
- [ ] Target document exists
- [ ] Target section exists
- [ ] Reference text matches target
- [ ] Link format correct
- [ ] Version compatibility maintained

---

> **Document Status**: Complete v1.0  
> **Next Review**: 2025-02-15  
> **Stakeholders**: System Architect, IV&V Lead, Product Owner
