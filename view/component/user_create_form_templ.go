// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package component

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func CreateUserForm(code string, role string) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<li id=\"create-user-form\" class=\"flex flex-col py-5\"><form hx-post=\"/api/user\" hx-target=\"#create-user-form\" hx-swap=\"outerHTML\" hx-target-error=\"#global-alert\" hx-indicator=\"#loading-overlay\" class=\"w-full grid grid-cols-1 sm:grid-cols-2 gap-4 mb-0\"><!-- Hidden Facility ID --><input type=\"hidden\" id=\"facility_code\" name=\"facility_code\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(code)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/user_create_form.templ`, Line: 11, Col: 16}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\"><!-- First Name --><div class=\"w-full\"><label for=\"first_name\" class=\"block text-sm/6 font-medium text-gray-900\">First Name</label> <input id=\"first_name\" name=\"first_name\" type=\"text\" required placeholder=\"John\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><!-- Last Name --><div class=\"w-full\"><label for=\"last_name\" class=\"block text-sm/6 font-medium text-gray-900\">Last Name</label> <input id=\"last_name\" name=\"last_name\" type=\"text\" required placeholder=\"Doe\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><!-- Initials --><div class=\"w-full\"><label for=\"initials\" class=\"block text-sm/6 font-medium text-gray-900\">Initials</label> <input id=\"initials\" name=\"initials\" type=\"text\" required maxlength=\"10\" placeholder=\"JD\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><!-- Email --><div class=\"w-full\"><label for=\"email\" class=\"block text-sm/6 font-medium text-gray-900\">Email</label> <input id=\"email\" name=\"email\" type=\"email\" required placeholder=\"john.doe@example.com\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><!-- Password --><div class=\"w-full\"><label for=\"password\" class=\"block text-sm/6 font-medium text-gray-900\">Password</label> <input id=\"password\" name=\"password\" type=\"password\" required minlength=\"8\" class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"></div><!-- Role --><div class=\"w-full\"><label for=\"role\" class=\"block text-sm/6 font-medium text-gray-900\">Role</label> <select id=\"role\" name=\"role\" required class=\"block w-full ring-1 ring-inset ring-gray-300 py-1.5 pl-1 rounded-md border-gray-300 shadow-sm focus:border-picton-blue-500 focus:ring-picton-blue-500 sm:text-sm\"><option value=\"\">Select a role</option> <option selected value=\"user\">User</option> <option value=\"admin\">Admin</option> ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if role == "super" {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<option value=\"super\">Super Admin</option>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</select></div><!-- Buttons --><div class=\"sm:col-span-2 flex justify-end gap-x-6 mt-4\"><button @click.prevent=\"$el.closest(&#39;li&#39;).innerHTML = &#39;&#39;\" class=\"text-sm/6 font-semibold text-gray-900\">Cancel</button> <button type=\"submit\" class=\"text-picton-blue-600 hover:text-picton-blue-900\">Save</button></div></form></li>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
