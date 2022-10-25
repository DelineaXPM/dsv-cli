package utils

func GetExecStatus(err error) int {
	if err == nil || err.Error() == "" {
		return 0
	}
	return 1
}
