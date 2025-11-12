// Package auth implements authentication and authorization for the Radio Control Container.
//
// The auth package validates bearer tokens and enforces scopes for radio operations,
// supporting radio:read and radio:control permissions.
//
// Architecture References:
//   - Architecture §14.1: Security and privacy requirements
//   - OpenAPI §3: Authentication specifications
package auth
