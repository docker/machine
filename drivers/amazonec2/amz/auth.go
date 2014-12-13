package amz

type Auth struct {
	AccessKey, SecretKey string
}

func GetAuth(accessKey, secretKey string) Auth {
	return Auth{accessKey, secretKey}
}
