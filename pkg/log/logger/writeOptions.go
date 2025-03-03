package logger

type WriteTo struct {
	//file/console/others
	Name string `json:"name"`
	//log level
	Level string `json:"level"`
	//params
	Args interface{} `json:"args"`
}

type File struct {
	//文件路径
	Path string `json:"path" default:"temp/logs"`
	//允许大小
	Cap uint `json:"cap"`
	//文件扩展名
	Suffix string `json:"suffix"`
	//文件大小限制 MB
	MaxSize int `json:"maxSize"`
	//最大保留文件数量
	MaxBackups int `json:"maxBackups"`
	//最大保留天数
	MaxAge int `json:"maxAge"`
	//是否压缩处理
	Compress bool `json:"compress"`
}
