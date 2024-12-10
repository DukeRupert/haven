// internal/model/params/facility.go
package params

type CreateFacilityParams struct {
	Name string `json:"name" form:"name"`
	Code string `json:"code" form:"code"`
}

type UpdateFacilityParams struct {
	Name string `json:"name" form:"name"`
	Code string `json:"code" form:"code"`
}
