// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package component

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"fmt"
	"github.com/DukeRupert/haven/types"
	"strconv"
)

func UpdateUserForm(user types.User, auth types.AuthContext) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<form hx-post=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(fmt.Sprintf("/api/user/%d", user.ID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/user_update_form.templ`, Line: 10, Col: 53}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" hx-target=\"this\" hx-swap=\"outerHTML\" hx-target-error=\"#global-alert\" hx-indicator=\"#loading-overlay\" class=\"w-full grid grid-cols-1 sm:grid-cols-2 gap-4 mb-0\"><input type=\"hidden\" name=\"facility_id\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(strconv.Itoa(user.FacilityID))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/user_update_form.templ`, Line: 11, Col: 79}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><div class=\"w-full\"><label for=\"first_name\" class=\"block text-sm/6 font-medium text-gray-900\">First Name</label> <input id=\"first_name\" name=\"first_name\" type=\"text\" required value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var4 string
		templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(user.FirstName)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/user_update_form.templ`, Line: 19, Col: 26}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" placeholder=\"John\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><div class=\"w-full\"><label for=\"last_name\" class=\"block text-sm/6 font-medium text-gray-900\">Last Name</label> <input id=\"last_name\" name=\"last_name\" type=\"text\" required value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var5 string
		templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(user.LastName)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/user_update_form.templ`, Line: 31, Col: 25}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" placeholder=\"Doe\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><div class=\"w-full\"><label for=\"initials\" class=\"block text-sm/6 font-medium text-gray-900\">Initials</label> <input id=\"initials\" name=\"initials\" type=\"text\" required value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var6 string
		templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(user.Initials)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/user_update_form.templ`, Line: 43, Col: 25}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" maxlength=\"2\" placeholder=\"JD\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><div class=\"w-full\"><label for=\"email\" class=\"block text-sm/6 font-medium text-gray-900\">Email</label> <input id=\"email\" name=\"email\" type=\"email\" required value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var7 string
		templ_7745c5c3_Var7, templ_7745c5c3_Err = templ.JoinStringErrs(user.Email)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/user_update_form.templ`, Line: 56, Col: 22}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var7))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" placeholder=\"john.doe@example.com\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if auth.Role != "user" {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"w-full\"><label for=\"role\" class=\"block text-sm/6 font-medium text-gray-900\">Role</label> <select id=\"role\" name=\"role\" required class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"><option value=\"\">Select a role</option> <option value=\"user\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if user.Role == "user" {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">User</option> <option value=\"admin\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if user.Role == "admin" {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Admin</option> ")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if auth.Role == "super" {
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<option value=\"super\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				if user.Role == "super" {
					_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
					if templ_7745c5c3_Err != nil {
						return templ_7745c5c3_Err
					}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Super Admin</option>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</select></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"sm:col-span-2 flex justify-end gap-x-6 mt-4\"><button type=\"submit\" class=\"text-picton-blue-600 hover:text-picton-blue-900\">Save</button></div></form>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
