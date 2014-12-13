package amz

type IpPermission struct {
	Protocol string
	FromPort int
	ToPort   int
	IpRange  string
}
