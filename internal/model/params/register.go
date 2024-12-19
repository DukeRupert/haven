package params

type RegisterParams struct {
	FacilityCode string `form:"facility_code"`
	Initials     string `form:"initials"`
	Email        string `form:"email"`
	Token        string `form:"token"`
}
