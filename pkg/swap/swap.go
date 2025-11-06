package swap

import "encoding/json"

// SwapTo
//
//	@Description: 转换
//	@param request
//	@param category
//	@return err
func SwapTo(request, category interface{}) (err error) {
	dataByte, err := json.Marshal(request)
	if err != nil {
		return
	}
	return json.Unmarshal(dataByte, category)
}
