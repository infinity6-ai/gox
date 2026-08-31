package cryptzjwt

import (
	"time"

	"github.com/google/uuid"
)

type builderLast interface {
	Build() (*JWTToken, error)
}

// builderSetKey is the builder step for setting the Header Key.
type builderSetKey interface {
	SetKey(key string) builderLast
}

// builderSetHeaderTyp is the builder step for setting the Header Type.
type builderSetHeaderTyp interface {
	SetHeaderType(typ string) builderSetKey
}

// builderSetAlg is the builder step for setting the Algorithm.
type builderSetAlg interface {
	SetAlg(alg string) builderSetHeaderTyp
	BuildPayload() (*JWTPayload, error)
}

// builderSetIdp is the builder step for setting the Identity Provider.
type builderSetIdp interface {
	SetIdentityProvider(idp string) builderSetAlg
}

// builderSetExt is the builder step for setting the Extension.
type builderSetExt interface {
	SetExtension(ext string) builderSetIdp
	SetBinaryExtension(bxt []byte) builderSetIdp
	SetBothExtentions(ext string, bxt []byte) builderSetIdp
}

// builderSetSubject is the builder step for setting the Subject.
type builderSetSubject interface {
	SetSubject(sub string) builderSetExt
}

// builderSetAudience is the builder step for setting the Audience.
type builderSetAudience interface {
	SetAudience(aud string) builderSetSubject
}

// builderSetIssuer is the builder step for setting the Issuer.
type builderSetIssuer interface {
	SetIssuer(iss string) builderSetAudience
}

// builderSetExpiration is the builder step for setting the Expiration.
type builderSetExpiration interface {
	ExpiresAt(t time.Time) builderSetIssuer
	ExpiresAtEpoch(t int64) builderSetIssuer
	ExpiresIn(d time.Duration) builderSetIssuer
}

// builderSetIat is the builder step for setting the Issued At.
type builderSetIat interface {
	CreatedAt(t time.Time) builderSetExpiration
	CreatedAtEpoch(t int64) builderSetExpiration
	CreatedRightNow() builderSetExpiration
}

// builderSetPayloadTyp is the builder step for setting the Payload Type.
type builderSetPayloadTyp interface {
	SetPayloadType(typ string) builderSetIat
}

// builderSetJti is the builder step for setting the JWT ID.
type builderSetJti interface {
	SetJti(id string) builderSetPayloadTyp
	GenerateJti() builderSetPayloadTyp
}

type builderImpl struct {
	token *JWTToken
}

// SetBinaryExtension implements [builderSetExt].
func (b *builderImpl) SetBinaryExtension(bxt []byte) builderSetIdp {
	b.token.Payload.Bxt = bxt
	return b
}

// SetBothExtentions implements [builderSetExt].
func (b *builderImpl) SetBothExtentions(ext string, bxt []byte) builderSetIdp {
	b.SetBinaryExtension(bxt)
	b.SetExtension(ext)
	return b
}

// ExpiresAtEpoch implements [builderSetExpiration].
func (b *builderImpl) ExpiresAtEpoch(t int64) builderSetIssuer {
	b.token.Payload.Exp = t
	return b
}

// CreatedAtEpoch implements [builderSetIat].
func (b *builderImpl) CreatedAtEpoch(t int64) builderSetExpiration {
	b.token.Payload.Iat = t
	return b
}

func (b *builderImpl) BuildPayload() (*JWTPayload, error) {
	err := basicVerify(b.token)
	return b.token.Payload, err
}

// NewToken starts the JWT building process.
func NewToken() builderSetJti {
	builder := &builderImpl{
		token: &JWTToken{
			Header:  &JWTHeader{},
			Payload: &JWTPayload{},
		},
	}
	return builder
}

func (b *builderImpl) SetJti(id string) builderSetPayloadTyp {
	b.token.Payload.Jti = id
	return b
}

func (b *builderImpl) GenerateJti() builderSetPayloadTyp {
	return b.SetJti(uuid.NewString())
}

func (b *builderImpl) SetPayloadType(typ string) builderSetIat {
	b.token.Payload.Typ = typ
	return b
}

func (b *builderImpl) CreatedAt(t time.Time) builderSetExpiration {
	return b.CreatedAtEpoch(t.UTC().UnixMilli())
}

func (b *builderImpl) CreatedRightNow() builderSetExpiration {
	return b.CreatedAt(time.Now().UTC())
}

func (b *builderImpl) ExpiresAt(t time.Time) builderSetIssuer {
	return b.ExpiresAtEpoch(t.UTC().UnixMilli())
}

func (b *builderImpl) ExpiresIn(d time.Duration) builderSetIssuer {
	iat := time.UnixMilli(b.token.Payload.Iat).UTC()
	return b.ExpiresAt(iat.Add(d))
}

func (b *builderImpl) SetIssuer(iss string) builderSetAudience {
	b.token.Payload.Iss = iss
	return b
}

func (b *builderImpl) SetAudience(aud string) builderSetSubject {
	b.token.Payload.Aud = aud
	return b
}

func (b *builderImpl) SetSubject(sub string) builderSetExt {
	b.token.Payload.Sub = sub
	return b
}

func (b *builderImpl) SetExtension(ext string) builderSetIdp {
	b.token.Payload.Ext = ext
	return b
}

func (b *builderImpl) SetIdentityProvider(idp string) builderSetAlg {
	b.token.Payload.Idp = idp
	return b
}

func (b *builderImpl) SetAlg(alg string) builderSetHeaderTyp {
	b.token.Header.Alg = alg
	return b
}

func (b *builderImpl) SetHeaderType(typ string) builderSetKey {
	b.token.Header.Typ = typ
	return b
}

func (b *builderImpl) SetKey(key string) builderLast {
	b.token.Header.Key = key
	return b
}

func (b *builderImpl) Build() (*JWTToken, error) {
	err := basicVerify(b.token)
	return b.token, err
}
