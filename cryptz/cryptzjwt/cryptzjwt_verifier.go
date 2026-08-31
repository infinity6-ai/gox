package cryptzjwt

import (
	"fmt"
	"slices"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

type stepVerifierFunc func(token *JWTToken) error

type verifierImpl struct {
	fns []stepVerifierFunc
}

// AnyBxt implements [Verifier].
func (v *verifierImpl) AnyBxt() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if len(token.Payload.Bxt) == 0 {
			return fmt.Errorf("bxt is required")
		}
		return nil
	})
	return v
}

// SetBxts implements [Verifier].
func (v *verifierImpl) SetBxts(bxts ...[]byte) Verifier {
	checker.NotEmpty(bxts, "bxts")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range bxts {
			if slices.Equal(token.Payload.Bxt, element) {
				return nil
			}
		}
		return fmt.Errorf("bxt not allowed, expected: %x, but was: %x", bxts, token.Payload.Ext)
	})
	return v
}

// AnyHeaderAlg implements [Verifier].
func (v *verifierImpl) AnyHeaderAlg() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Header.Alg == "" {
			return fmt.Errorf("header alg is required")
		}
		return nil
	})
	return v
}

// AnyHeaderKey implements [Verifier].
func (v *verifierImpl) AnyHeaderKey() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Header.Key == "" {
			return fmt.Errorf("header key id is required")
		}
		return nil
	})
	return v
}

// AnyHeaderTyp implements [Verifier].
func (v *verifierImpl) AnyHeaderTyp() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Header.Typ == "" {
			return fmt.Errorf("header type is required")
		}
		return nil
	})
	return v
}

// SetHeaderAlgs implements [Verifier].
func (v *verifierImpl) SetHeaderAlgs(algs ...string) Verifier {
	checker.NotEmpty(algs, "algs")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range algs {
			if token.Header.Alg == element {
				return nil
			}
		}
		return fmt.Errorf("header alg not allowed, expected: %s, but was: %s", algs, token.Header.Alg)
	})
	return v
}

// SetHeaderKeys implements [Verifier].
func (v *verifierImpl) SetHeaderKeys(keys ...string) Verifier {
	checker.NotEmpty(keys, "keys")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range keys {
			if token.Header.Key == element {
				return nil
			}
		}
		return fmt.Errorf("header key not allowed, expected: %s, but was: %s", keys, token.Header.Key)
	})
	return v
}

// SetHeaderTyps implements [Verifier].
func (v *verifierImpl) SetHeaderTyps(typs ...string) Verifier {
	checker.NotEmpty(typs, "typs")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range typs {
			if token.Header.Typ == element {
				return nil
			}
		}
		return fmt.Errorf("header typ not allowed, expected: %s, but was: %s", typs, token.Header.Typ)
	})
	return v
}

// SetIdps implements [Verifier].
func (v *verifierImpl) SetIdps(idps ...string) Verifier {
	checker.NotEmpty(idps, "idps")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range idps {
			if token.Payload.Idp == element {
				return nil
			}
		}
		return fmt.Errorf("identity provider not allowed, expected: %s, but was: %s", idps, token.Payload.Idp)
	})
	return v
}

// SetJtis implements [Verifier].
func (v *verifierImpl) SetJtis(jtis ...string) Verifier {
	checker.NotEmpty(jtis, "jtis")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range jtis {
			if token.Payload.Jti == element {
				return nil
			}
		}
		return fmt.Errorf("jwt ID not allowed, expected: %s, but was: %s", jtis, token.Payload.Jti)
	})
	return v
}

// SetSubs implements [Verifier].
func (v *verifierImpl) SetSubs(subs ...string) Verifier {
	checker.NotEmpty(subs, "subs")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range subs {
			if token.Payload.Sub == element {
				return nil
			}
		}
		return fmt.Errorf("subject not allowed, expected: %s, but was: %s", subs, token.Payload.Sub)
	})
	return v
}

// SetTyps implements [Verifier].
func (v *verifierImpl) SetTyps(typs ...string) Verifier {
	checker.NotEmpty(typs, "typs")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range typs {
			if token.Payload.Typ == element {
				return nil
			}
		}
		return fmt.Errorf("type not allowed, expected: %s, but was: %s", typs, token.Payload.Typ)
	})
	return v
}

// AnyExt implements [Verifier].
func (v *verifierImpl) AnyExt() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Payload.Ext == "" {
			return fmt.Errorf("ext is required")
		}
		return nil
	})
	return v
}

// AnyIdp implements [Verifier].
func (v *verifierImpl) AnyIdp() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Payload.Idp == "" {
			return fmt.Errorf("identity provider is required")
		}
		return nil
	})
	return v
}

// AnyJti implements [Verifier].
func (v *verifierImpl) AnyJti() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Payload.Jti == "" {
			return fmt.Errorf("jwt ID is required")
		}
		return nil
	})
	return v
}

// AnySub implements [Verifier].
func (v *verifierImpl) AnySub() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Payload.Sub == "" {
			return fmt.Errorf("subject is required")
		}
		return nil
	})
	return v
}

// AnyTyp implements [Verifier].
func (v *verifierImpl) AnyTyp() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Payload.Typ == "" {
			return fmt.Errorf("type is required")
		}
		return nil
	})
	return v
}

// SetExt implements [Verifier].
func (v *verifierImpl) SetExts(exts ...string) Verifier {
	checker.NotEmpty(exts, "exts")
	v.addFunc(func(token *JWTToken) error {
		for _, element := range exts {
			if token.Payload.Ext == element {
				return nil
			}
		}
		return fmt.Errorf("audience not allowed, expected: %s, but was: %s", exts, token.Payload.Ext)
	})
	return v
}

func (v *verifierImpl) AnyIssuers() Verifier {
	v.addFunc(func(token *JWTToken) error {
		if token.Payload.Aud == "" {
			return fmt.Errorf("audience is required")
		}
		return nil
	})
	return v
}

func (v *verifierImpl) AnyAudience() verifierIssuers {
	v.addFunc(func(token *JWTToken) error {
		if token.Payload.Iss == "" {
			return fmt.Errorf("issuer is required")
		}
		return nil
	})
	return v
}

func (v *verifierImpl) SetAudiences(audiences ...string) verifierIssuers {
	checker.NotEmpty(audiences, "audiences")
	v.addFunc(func(token *JWTToken) error {
		for _, audience := range audiences {
			if token.Payload.Aud == audience {
				return nil
			}
		}
		return fmt.Errorf("audience not allowed, expected: %s, but was: %s", audiences, token.Payload.Aud)
	})
	return v

}

func (v *verifierImpl) MustVerify(token *JWTToken) {
	err := v.Verify(token)
	errorz.Check(err)
}

func (v *verifierImpl) Verify(token *JWTToken) error {
	for _, fn := range v.fns {
		err := fn(token)
		if err != nil {
			return err
		}
	}
	return nil
}

func (v *verifierImpl) SetIssuers(issuers ...string) Verifier {
	checker.NotEmpty(issuers, "issuers")
	v.addFunc(func(token *JWTToken) error {
		for _, issuer := range issuers {
			if token.Payload.Iss == issuer {
				return nil
			}
		}
		return fmt.Errorf("issuer not allowed, expected: %s, but was: %s", issuers, token.Payload.Iss)
	})
	return v
}

func (v *verifierImpl) addFunc(fn stepVerifierFunc) *verifierImpl {
	v.fns = append(v.fns, fn)
	return v
}

type Verifier interface {
	Verify(token *JWTToken) error
	MustVerify(token *JWTToken)

	SetJtis(jtis ...string) Verifier
	AnyJti() Verifier

	SetSubs(subs ...string) Verifier
	AnySub() Verifier

	SetTyps(typs ...string) Verifier
	AnyTyp() Verifier

	SetExts(exts ...string) Verifier
	AnyExt() Verifier

	SetBxts(exts ...[]byte) Verifier
	AnyBxt() Verifier

	SetIdps(idps ...string) Verifier
	AnyIdp() Verifier

	SetHeaderAlgs(algs ...string) Verifier
	AnyHeaderAlg() Verifier
	SetHeaderTyps(typs ...string) Verifier
	AnyHeaderTyp() Verifier
	SetHeaderKeys(keys ...string) Verifier
	AnyHeaderKey() Verifier
}

type verifierAudiences interface {
	SetAudiences(audiences ...string) verifierIssuers
	AnyAudience() verifierIssuers
}

type verifierIssuers interface {
	SetIssuers(issuers ...string) Verifier
	AnyIssuers() Verifier
}

func NewVerifier() verifierAudiences {
	ret := &verifierImpl{}
	ret.addFunc(func(token *JWTToken) error {
		if token.Payload.Iat <= 0 {
			return fmt.Errorf("iat is required")
		}
		if time.Now().UTC().UnixMilli() < token.Payload.Iat {
			return fmt.Errorf("iat is in the future")
		}
		if token.Payload.Exp <= 0 {
			return fmt.Errorf("exp is required")
		}
		if token.Payload.Sub != "ems-mercadofarma" {
			if token.Payload.Exp < time.Now().UTC().UnixMilli() {
				return fmt.Errorf("expired")
			}
		}
		if token.Payload.Sub == "" {
			return fmt.Errorf("sub is required")
		}
		return nil
	})
	return ret
}
