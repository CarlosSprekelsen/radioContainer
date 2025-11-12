# Internal Package Documentation Policy

## Documentation Standards

All internal packages follow Go documentation conventions with architecture traceability:

### Package Documentation (`doc.go`)
- Single paragraph component summary
- Architecture References section with 3-5 bullet points
- No inline citations in individual files

### Function Documentation
- Start with function name
- Describe what, not how
- Include error conditions where relevant
- Keep lines ≤100 characters
- Complete sentences with proper punctuation

### Test Documentation
- Minimal documentation (godoc skips test files)
- Clear, concise test descriptions
- No verbose architecture citations

## Architecture References Format
```
// Architecture References:
//   - Architecture §X: Component description
//   - CB-TIMING §Y: Timing constraints
//   - OpenAPI §Z: API specifications
```

## Anti-Patterns to Avoid
- ❌ Inline "Source:" and "Quote:" citations
- ❌ Duplicate architecture references across files
- ❌ Verbose test documentation
- ❌ Requirements restatement in function docs

## Quality Gates
- All exported symbols documented
- Package docs reference architecture sections
- No duplicate citations
- Clean `go doc` output
