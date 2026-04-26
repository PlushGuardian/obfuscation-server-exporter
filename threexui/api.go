package threexui

import "encoding/json"

type apiResponse struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj"`
}

type serverStatusResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Obj     struct {
		Xray struct {
			Version string `json:"version"`
		} `json:"xray"`
		AppStats struct {
			Threads int   `json:"threads"`
			Mem     int64 `json:"mem"`
			Uptime  int64 `json:"uptime"`
		} `json:"appStats"`
	} `json:"obj"`
}

type getInboundsResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Obj     []struct {
		ID          int    `json:"id"`
		Remark      string `json:"remark"`
		Up          int64  `json:"up"`
		Down        int64  `json:"down"`
		ClientStats []struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
			Up    int64  `json:"up"`
			Down  int64  `json:"down"`
		} `json:"clientStats"`
	} `json:"obj"`
}
