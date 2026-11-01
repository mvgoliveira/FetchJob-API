package errors

import "errors"

var (
	ErrFetchingSecretFromSecretManager        = errors.New("could not fetch secret from secret manager")
	ErrGcpVariableNotSet                      = errors.New("the environment variable GOOGLE_CLOUD_PROJECT value not set")
	ErrInstantiatingSecretManagerClient       = errors.New("could not instantiate Secret Manager client")
	ErrProcessingEnvironmentIntoConfiguration = errors.New("could not process configuration from environment")
	ErrUnmarshallingSecretIntoConfiguration   = errors.New("could not unmarshall secret into configuration")
)
