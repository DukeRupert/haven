// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package auth

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import "github.com/DukeRupert/haven/view/layout"

func Login() templ.Component {
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
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
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
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"flex min-h-full flex-col justify-center px-6 py-12 lg:px-8\"><div class=\"sm:mx-auto sm:w-full sm:max-w-sm\"><img class=\"mx-auto h-16 w-16\" src=\"static/logo.svg\" alt=\"Haven\"><h2 class=\"mt-10 text-center text-2xl/9 font-bold tracking-tight text-gray-900\">Sign in to your account</h2></div><div class=\"mt-10 sm:mx-auto sm:w-full sm:max-w-sm\"><form class=\"space-y-6\" action=\"/login\" method=\"POST\"><div><label for=\"email\" class=\"block text-sm/6 font-medium text-gray-900\">Email address</label><div class=\"mt-2\"><input id=\"email\" name=\"email\" type=\"email\" autocomplete=\"email\" required class=\"block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm/6\"></div></div><div><div class=\"flex items-center justify-between\"><label for=\"password\" class=\"block text-sm/6 font-medium text-gray-900\">Password</label><div class=\"text-sm\"><a href=\"#\" class=\"font-semibold text-indigo-600 hover:text-indigo-500\">Forgot password?</a></div></div><div class=\"mt-2\"><input id=\"password\" name=\"password\" type=\"password\" autocomplete=\"current-password\" required minlength=\"8\" class=\"block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm/6\"></div></div><div><button type=\"submit\" class=\"flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm/6 font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600\">Sign in</button></div></form><p class=\"mt-10 text-center text-sm/6 text-gray-500\">Not a member? <a href=\"#\" class=\"font-semibold text-indigo-600 hover:text-indigo-500\">Start a 14 day free trial</a></p></div></div><script>\ndocument.querySelector('form').addEventListener('submit', async function(e) {\n    e.preventDefault();\n    \n    const formData = {\n        email: document.getElementById('email').value,\n        password: document.getElementById('password').value\n    };\n\n    try {\n        const response = await fetch('/login', {\n            method: 'POST',\n            headers: {\n                'Content-Type': 'application/json',\n            },\n            body: JSON.stringify(formData)\n        });\n\n        if (!response.ok) {\n            const errorData = await response.json();\n            throw new Error(errorData.message || 'Login failed');\n        }\n\n        const data = await response.json();\n        console.log('Login successful:', data);\n        \n        // Redirect after successful login\n        window.location.href = '/app/dashboard'; // Adjust the redirect URL as needed\n        \n    } catch (error) {\n        console.error('Login error:', error);\n        // Handle error (show error message to user)\n        alert(error.message || 'Login failed. Please try again.');\n    }\n});\n</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return templ_7745c5c3_Err
		})
		templ_7745c5c3_Err = layout.BaseLayout().Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
