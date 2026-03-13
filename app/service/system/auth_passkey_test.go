package system

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	systemModel "github.com/seakee/go-api/app/model/system"
	"github.com/seakee/go-api/app/pkg/e"
	"gorm.io/gorm"
)

func TestAuthService_DeletePasskey(t *testing.T) {
	t.Run("empty passkey id returns explicit error", func(t *testing.T) {
		svc := &authService{}

		errCode, err := svc.DeletePasskey(context.Background(), 1, 0)
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.PasskeyCanNotBeNull {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.PasskeyCanNotBeNull)
		}
	})

	t.Run("reject delete when passkey is last login method", func(t *testing.T) {
		deleteCalled := false

		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11)
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.LastLoginMethodCannotBeRemoved {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.LastLoginMethodCannotBeRemoved)
		}
		if deleteCalled {
			t.Fatalf("DeletePasskey() should not call DeleteByIDAndUserID()")
		}
	})

	t.Run("delete passkey succeeds when other login method exists", func(t *testing.T) {
		deleteCalled := false

		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1, Password: "hashed"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11)
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if !deleteCalled {
			t.Fatalf("DeletePasskey() should call DeleteByIDAndUserID()")
		}
	})

	t.Run("delete passkey succeeds when another passkey remains", func(t *testing.T) {
		deleteCalled := false

		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 2, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					deleteCalled = true
					return nil
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11)
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.SUCCESS {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.SUCCESS)
		}
		if !deleteCalled {
			t.Fatalf("DeletePasskey() should call DeleteByIDAndUserID()")
		}
	})

	t.Run("delete maps record not found to passkey not found", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Status: 1, Password: "hashed"}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				DetailByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) (*systemModel.UserPasskey, error) {
					return &systemModel.UserPasskey{Model: gorm.Model{ID: id}, UserID: userID}, nil
				},
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
				DeleteByIDAndUserIDFunc: func(ctx context.Context, id, userID uint) error {
					return gorm.ErrRecordNotFound
				},
			},
		}

		errCode, err := svc.DeletePasskey(context.Background(), 7, 11)
		if err != nil {
			t.Fatalf("DeletePasskey() error = %v, want nil", err)
		}
		if errCode != e.PasskeyCredentialNotFound {
			t.Fatalf("DeletePasskey() errCode = %d, want %d", errCode, e.PasskeyCredentialNotFound)
		}
	})
}

func TestAuthService_ListPasskeys(t *testing.T) {
	t.Run("returns user not found when parent user is missing", func(t *testing.T) {
		listCalled := false
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return nil, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserPasskey, error) {
					listCalled = true
					return nil, nil
				},
			},
		}

		list, errCode, err := svc.ListPasskeys(context.Background(), 9)
		if err != nil {
			t.Fatalf("ListPasskeys() error = %v, want nil", err)
		}
		if errCode != e.UserNotFound {
			t.Fatalf("ListPasskeys() errCode = %d, want %d", errCode, e.UserNotFound)
		}
		if list != nil {
			t.Fatalf("ListPasskeys() list = %+v, want nil", list)
		}
		if listCalled {
			t.Fatalf("ListPasskeys() should not call ListByUserID() when user is missing")
		}
	})
}

func TestAuthService_loginMethodCount(t *testing.T) {
	t.Run("counts password identity and passkey methods", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Password: "hashed"}, nil
				},
			},
			identityRepo: &mockUserIdentityRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserIdentity, error) {
					return []systemModel.UserIdentity{{Model: gorm.Model{ID: 1}, UserID: userID}}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 1, nil
				},
			},
		}

		count, err := svc.loginMethodCount(context.Background(), 9)
		if err != nil {
			t.Fatalf("loginMethodCount() error = %v", err)
		}
		if count != 3 {
			t.Fatalf("loginMethodCount() count = %d, want 3", count)
		}
	})

	t.Run("counts each identity and passkey credential individually", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}, Password: "hashed"}, nil
				},
			},
			identityRepo: &mockUserIdentityRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserIdentity, error) {
					return []systemModel.UserIdentity{
						{Model: gorm.Model{ID: 1}, UserID: userID},
						{Model: gorm.Model{ID: 2}, UserID: userID},
					}, nil
				},
			},
			passkeyRepo: &mockUserPasskeyRepo{
				CountByUserIDFunc: func(ctx context.Context, userID uint) (int64, error) {
					return 2, nil
				},
			},
		}

		count, err := svc.loginMethodCount(context.Background(), 9)
		if err != nil {
			t.Fatalf("loginMethodCount() error = %v", err)
		}
		if count != 5 {
			t.Fatalf("loginMethodCount() count = %d, want 5", count)
		}
	})

	t.Run("returns zero when user does not exist", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return nil, nil
				},
			},
		}

		count, err := svc.loginMethodCount(context.Background(), 9)
		if err != nil {
			t.Fatalf("loginMethodCount() error = %v", err)
		}
		if count != 0 {
			t.Fatalf("loginMethodCount() count = %d, want 0", count)
		}
	})

	t.Run("returns identity repo error", func(t *testing.T) {
		svc := &authService{
			userRepo: &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*systemModel.User, error) {
					return &systemModel.User{Model: gorm.Model{ID: id}}, nil
				},
			},
			identityRepo: &mockUserIdentityRepo{
				ListByUserIDFunc: func(ctx context.Context, userID uint) ([]systemModel.UserIdentity, error) {
					return nil, errors.New("list identities failed")
				},
			},
		}

		_, err := svc.loginMethodCount(context.Background(), 9)
		if err == nil {
			t.Fatalf("loginMethodCount() error = nil, want non-nil")
		}
	})
}

func TestParseStoredCredential(t *testing.T) {
	t.Run("parses json credential first", func(t *testing.T) {
		stored := webauthn.Credential{
			ID:        []byte{1, 2, 3},
			PublicKey: []byte{4, 5, 6},
			Transport: []protocol.AuthenticatorTransport{protocol.AuthenticatorTransport("usb")},
			Authenticator: webauthn.Authenticator{
				SignCount: 8,
				AAGUID:    []byte{9, 10},
			},
		}
		raw, err := json.Marshal(stored)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		credential, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialPublicKey: string(raw),
		})
		if !ok {
			t.Fatalf("parseStoredCredential() ok = false, want true")
		}
		if string(credential.ID) != string(stored.ID) {
			t.Fatalf("parseStoredCredential() credential.ID = %v, want %v", credential.ID, stored.ID)
		}
		if string(credential.PublicKey) != string(stored.PublicKey) {
			t.Fatalf("parseStoredCredential() credential.PublicKey = %v, want %v", credential.PublicKey, stored.PublicKey)
		}
		if credential.Authenticator.SignCount != stored.Authenticator.SignCount {
			t.Fatalf("parseStoredCredential() credential.Authenticator.SignCount = %d, want %d",
				credential.Authenticator.SignCount, stored.Authenticator.SignCount)
		}
	})

	t.Run("falls back to legacy credential fields", func(t *testing.T) {
		credential, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialID:        encodeBytesToBase64URL([]byte{11, 12}),
			CredentialPublicKey: encodeBytesToBase64URL([]byte{13, 14}),
			SignCount:           10,
			AAGUID:              encodeAAGUID([]byte{15, 16}),
			TransportsJSON:      `["usb","internal"]`,
		})
		if !ok {
			t.Fatalf("parseStoredCredential() ok = false, want true")
		}
		if string(credential.ID) != string([]byte{11, 12}) {
			t.Fatalf("parseStoredCredential() credential.ID = %v, want %v", credential.ID, []byte{11, 12})
		}
		if string(credential.PublicKey) != string([]byte{13, 14}) {
			t.Fatalf("parseStoredCredential() credential.PublicKey = %v, want %v", credential.PublicKey, []byte{13, 14})
		}
		if credential.Authenticator.SignCount != 10 {
			t.Fatalf("parseStoredCredential() credential.Authenticator.SignCount = %d, want 10", credential.Authenticator.SignCount)
		}
		if string(credential.Authenticator.AAGUID) != string([]byte{15, 16}) {
			t.Fatalf("parseStoredCredential() credential.Authenticator.AAGUID = %v, want %v", credential.Authenticator.AAGUID, []byte{15, 16})
		}
		if len(credential.Transport) != 2 {
			t.Fatalf("parseStoredCredential() len(credential.Transport) = %d, want 2", len(credential.Transport))
		}
	})

	t.Run("returns false when credential id is invalid", func(t *testing.T) {
		_, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialID:        "invalid***",
			CredentialPublicKey: "",
		})
		if ok {
			t.Fatalf("parseStoredCredential() ok = true, want false")
		}
	})

	t.Run("returns false when legacy public key is invalid", func(t *testing.T) {
		_, ok := parseStoredCredential(systemModel.UserPasskey{
			CredentialID:        encodeBytesToBase64URL([]byte{1, 2}),
			CredentialPublicKey: "invalid***",
		})
		if ok {
			t.Fatalf("parseStoredCredential() ok = true, want false")
		}
	})
}

func TestMarshalPasskeyCredentialPayload(t *testing.T) {
	credential := PasskeyCredential{
		ID: "credential-id",
		Response: map[string]any{
			"client_data_json":   "client-data",
			"authenticator_data": "auth-data",
			"user_handle":        "user-handle",
			"nested": map[string]any{
				"public_key":           "pk",
				"public_key_algorithm": -7,
			},
			"list": []any{
				map[string]any{"attestation_object": "attestation"},
			},
		},
	}

	raw, err := marshalPasskeyCredentialPayload(credential)
	if err != nil {
		t.Fatalf("marshalPasskeyCredentialPayload() error = %v", err)
	}

	var payload map[string]any
	if err = json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload["type"] != string(protocol.PublicKeyCredentialType) {
		t.Fatalf("payload.type = %v, want %s", payload["type"], protocol.PublicKeyCredentialType)
	}
	if payload["rawId"] != "credential-id" {
		t.Fatalf("payload.rawId = %v, want credential-id", payload["rawId"])
	}

	response, ok := payload["response"].(map[string]any)
	if !ok {
		t.Fatalf("payload.response type = %T, want map[string]any", payload["response"])
	}
	if _, has := response["client_data_json"]; has {
		t.Fatalf("response should not keep snake_case key client_data_json")
	}
	if response["clientDataJSON"] != "client-data" {
		t.Fatalf("response.clientDataJSON = %v, want client-data", response["clientDataJSON"])
	}
	if response["authenticatorData"] != "auth-data" {
		t.Fatalf("response.authenticatorData = %v, want auth-data", response["authenticatorData"])
	}
	if response["userHandle"] != "user-handle" {
		t.Fatalf("response.userHandle = %v, want user-handle", response["userHandle"])
	}

	nested, ok := response["nested"].(map[string]any)
	if !ok {
		t.Fatalf("response.nested type = %T, want map[string]any", response["nested"])
	}
	if nested["publicKey"] != "pk" {
		t.Fatalf("response.nested.publicKey = %v, want pk", nested["publicKey"])
	}
	if nested["publicKeyAlgorithm"] != float64(-7) {
		t.Fatalf("response.nested.publicKeyAlgorithm = %v, want -7", nested["publicKeyAlgorithm"])
	}

	list, ok := response["list"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("response.list = %v, want one element array", response["list"])
	}
	item, ok := list[0].(map[string]any)
	if !ok {
		t.Fatalf("response.list[0] type = %T, want map[string]any", list[0])
	}
	if item["attestationObject"] != "attestation" {
		t.Fatalf("response.list[0].attestationObject = %v, want attestation", item["attestationObject"])
	}
}

func TestMarshalPasskeyCredentialPayload_EmptyResponse(t *testing.T) {
	_, err := marshalPasskeyCredentialPayload(PasskeyCredential{
		ID:       "credential-id",
		Response: map[string]any{},
	})
	if err == nil {
		t.Fatalf("marshalPasskeyCredentialPayload() error = nil, want non-nil")
	}
}
