# Negative & Security Test Matrix

This document enumerates the negative-path and security-oriented E2E test scenarios for MaaS. Each scenario documents what is tested, the expected outcome, and where the test lives.

| # | Category | Scenario | Expected | Test | File |
|---|----------|----------|----------|------|------|
| 1 | Auth | No Authorization header | 401 | `test_no_auth_gets_401` | test_subscription.py |
| 2 | Auth | Invalid/garbage API key | 403 | `test_invalid_token_gets_403` | test_subscription.py |
| 3 | Auth | Wrong group on premium model | 403 | `test_wrong_group_gets_403` | test_subscription.py |
| 4 | Auth | Unauthenticated /v1/models | 401 | `test_unauthenticated_request_401` | test_models_endpoint.py |
| 5 | Header | API key ignores X-MaaS-Subscription | 200 (key-derived sub) | `test_api_key_ignores_subscription_header` | test_subscription.py |
| 6 | Header | Empty subscription header | Treated as missing | `test_empty_subscription_header_value` | test_models_endpoint.py |
| 7 | Header | Invalid subscription name | 403 | `test_invalid_subscription_header_403` | test_models_endpoint.py |
| 8 | Header | Access denied to subscription | 403 | `test_access_denied_to_subscription_403` | test_models_endpoint.py |
| 9 | Revocation | Revoked API key at gateway | 403 | `test_revoked_api_key_rejected` | test_api_keys.py |
| 10 | Revocation | Individually revoked keys rejected | 403 | `test_revoke_keys_rejected_at_gateway` | test_api_keys.py |
| 11 | Subscription | API key with deleted subscription | 403 | `test_api_key_with_deleted_subscription_403` | test_subscription.py |
| 12 | IDOR | Non-admin accessing other user's keys | 404 (not 403) | `test_non_admin_cannot_access_other_users_keys` | test_api_keys.py |
| 13 | Rate limit | Exhaustion returns 429 | 429 | `test_rate_limit_exhaustion_gets_429` | test_subscription.py |
| 14 | CWE-693/400 | TRLP persists during sub deletion | TRLP rebuilt | `test_trlp_persists_during_multi_subscription_deletion` | test_subscription.py |
| 15 | Header strip | Identity headers not forwarded | Unit test | `TestMaaSAuthPolicyReconciler_NoIdentityHeadersUpstream` | maasauthpolicy_controller_test.go |
| 16 | Isolation | Namespace scoping | 4 test classes | Multiple | test_namespace_scoping.py |
| 17 | Header spoofing | Client injects X-MaaS-Username, X-MaaS-Group, X-MaaS-Key-Id headers with valid API key | 200 with key-derived identity; spoofed headers ignored/stripped | `TestHeaderSpoofing::test_injected_identity_headers_ignored` | test_negative_security.py |
| 18 | Header spoofing | Client sends duplicate X-MaaS-Subscription headers | API key binding wins; duplicates don't bypass | `TestHeaderSpoofing::test_duplicate_subscription_headers_ignored` | test_negative_security.py |
| 19 | Auth bypass | Expired API key used for inference | 403 at gateway | `TestExpiredKeyRejection::test_expired_key_rejected_at_gateway` | test_negative_security.py |
| 20 | Missing resource | API key bound to sub A tries to access model B (not in sub A) | 403 at gateway | `TestCrossModelAccess::test_key_cannot_access_model_outside_subscription` | test_negative_security.py |
| 21 | Missing resource | MaaSAuthPolicy deleted while API key still active | 403 at gateway after propagation | `TestAuthPolicyRemoval::test_authpolicy_deletion_revokes_access` | test_negative_security.py |
| 22 | Missing resource | MaaSSubscription referencing non-existent MaaSModelRef | CR status not Active | `TestMissingModelRef::test_subscription_with_nonexistent_model_ref` | test_negative_security.py |
| 23 | Missing resource | MaaSAuthPolicy referencing non-existent MaaSModelRef | CR status not Active | `TestMissingModelRef::test_authpolicy_with_nonexistent_model_ref` | test_negative_security.py |
| 24 | Abuse | Inference with special chars in X-MaaS-Subscription header | 403 (not found); no injection | `TestHeaderAbuse::test_special_characters_in_subscription_header` | test_negative_security.py |
