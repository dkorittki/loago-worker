package client

type WrappableError interface {
	error
	Unwrap() error
}

type CertificateReadError struct {
	Err error
}

func (e *CertificateReadError) Error() string {
	return "cannot read certificate"
}

func (e *CertificateReadError) Unwrap() error {
	return e.Err
}

type CertificateDecodeError struct{}

func (e *CertificateDecodeError) Error() string {
	return "cannot decode certificate"
}

type ConnectError struct {
	Err error
}

func (e *ConnectError) Error() string {
	return "cannot create connection"
}

func (e *ConnectError) Unwrap() error {
	return e.Err
}

type DisconnectError struct {
	Err error
}

func (e *DisconnectError) Error() string {
	return "cannot close connection"
}

func (e *DisconnectError) Unwrap() error {
	return e.Err
}

type InvalidConnectionError struct {
	Err error
}

func (e *InvalidConnectionError) Error() string {
	return "invalid connection state"
}

func (e *InvalidConnectionError) Unwrap() error {
	return e.Err
}
