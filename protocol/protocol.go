package protocol

type LogGatherReport struct {
	FileName    string `json:"file_name"`
	LogInfoGzip []byte `json:"log_info"`
}

type LogGatherResp struct {
}
