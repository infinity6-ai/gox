package cryptzjwt

import "github.com/infinity6-ai/gox/cryptz/pb/cryptzjwtpb"

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Key string `json:"key"`
}

func (d *JWTHeader) FromProto(pb *cryptzjwtpb.JWTHeader) {
	d.Alg = pb.Alg
	d.Typ = pb.Typ
	d.Key = pb.Key
}

func (d *JWTHeader) ToProto() *cryptzjwtpb.JWTHeader {
	return &cryptzjwtpb.JWTHeader{
		Alg: d.Alg,
		Typ: d.Typ,
		Key: d.Key,
	}
}

type JWTPayload struct {
	Jti string `json:"jti,omitempty"`
	Iss string `json:"iss,omitempty"`
	Sub string `json:"sub,omitempty"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
	Aud string `json:"aud,omitempty"`
	Typ string `json:"typ,omitempty"`
	Ext string `json:"ext,omitempty"`
	Bxt []byte `json:"bxt,omitempty"`
	Idp string `json:"idp,omitempty"`
}

func (d *JWTPayload) FromProto(pb *cryptzjwtpb.JWTPayload) {
	d.Jti = pb.Jti
	d.Iss = pb.Iss
	d.Sub = pb.Sub
	d.Iat = pb.Iat
	d.Exp = pb.Exp
	d.Aud = pb.Aud
	d.Typ = pb.Typ
	d.Ext = pb.Ext
	d.Bxt = pb.Bxt
	d.Idp = pb.Idp
}

func (d *JWTPayload) ToProto() *cryptzjwtpb.JWTPayload {
	return &cryptzjwtpb.JWTPayload{
		Jti: d.Jti,
		Iss: d.Iss,
		Sub: d.Sub,
		Iat: d.Iat,
		Exp: d.Exp,
		Aud: d.Aud,
		Typ: d.Typ,
		Ext: d.Ext,
		Bxt: d.Bxt,
		Idp: d.Idp,
	}
}

type JWTToken struct {
	Header  *JWTHeader  `json:"header"`
	Payload *JWTPayload `json:"payload"`
}

func (d *JWTToken) FromProto(pb *cryptzjwtpb.JWTToken) {
	d.Header.FromProto(pb.Header)
	d.Payload.FromProto(pb.Payload)
}

func (d *JWTToken) ToProto() *cryptzjwtpb.JWTToken {
	return &cryptzjwtpb.JWTToken{
		Header:  d.Header.ToProto(),
		Payload: d.Payload.ToProto(),
	}
}
