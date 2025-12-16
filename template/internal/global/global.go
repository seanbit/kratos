package global

func GetEnv() string {
	return GetConfig().Env.String()
}

func GetServiceName() string {
	return GetConfig().Name
}
