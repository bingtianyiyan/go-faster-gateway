package env

type (
	Mode string
)

const (
	ModeDebug Mode = "debug"       //调试模式
	ModeDev   Mode = "development" //开发模式
	ModeTest  Mode = "staging"     //测试模式
	ModeProd  Mode = "production"  //生产模式
)

func (e Mode) String() string {
	return string(e)
}
