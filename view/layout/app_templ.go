// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package layout

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func AppLayout(pageTitle string) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"min-h-full\"><nav class=\"border-b border-gray-200 bg-white\" x-data=\"{ mobileMenuOpen: false, userMenuOpen: false }\"><div class=\"mx-auto max-w-7xl px-4 sm:px-6 lg:px-8\"><div class=\"flex h-16 justify-between\"><div class=\"flex\"><div class=\"flex shrink-0 items-center\"><img loading=\"lazy\" class=\"block h-8 w-auto lg:hidden\" src=\"/static/logo.svg\" alt=\"Haven, by Firefly Software\" height=\"32\" width=\"64\"> <img loading=\"lazy\" class=\"hidden h-8 w-auto lg:block\" src=\"/static/logo.svg\" alt=\"Haven, by Firefly Software\" height=\"32\" width=\"64\"></div><div class=\"hidden sm:-my-px sm:ml-6 sm:flex sm:space-x-8\"><a href=\"#\" class=\"inline-flex items-center border-b-2 border-red-500 px-1 pt-1 text-sm font-medium text-gray-900\" aria-current=\"page\">Dashboard</a> <a href=\"#\" class=\"inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700\">Team</a> <a href=\"#\" class=\"inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700\">Projects</a> <a href=\"#\" class=\"inline-flex items-center border-b-2 border-transparent px-1 pt-1 text-sm font-medium text-gray-500 hover:border-gray-300 hover:text-gray-700\">Calendar</a></div></div><div class=\"hidden sm:ml-6 sm:flex sm:items-center\"><button type=\"button\" class=\"relative rounded-full bg-white p-1 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2\"><span class=\"absolute -inset-1.5\"></span> <span class=\"sr-only\">View notifications</span> <svg height=\"24\" width=\"24\" class=\"h-6 w-6\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" aria-hidden=\"true\" data-slot=\"icon\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0\"></path></svg></button><!-- Profile dropdown --><div class=\"relative ml-3\" @click.away=\"userMenuOpen = false\"><div><button type=\"button\" @click=\"userMenuOpen = !userMenuOpen\" class=\"relative flex max-w-xs items-center rounded-full bg-white text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2\" id=\"user-menu-button\" :aria-expanded=\"userMenuOpen\"><span class=\"absolute -inset-1.5\"></span> <span class=\"sr-only\">Open user menu</span> <img height=\"32\" width=\"32\" class=\"h-8 w-8 rounded-full\" src=\"https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-1.2.1&amp;ixid=eyJhcHBfaWQiOjEyMDd9&amp;auto=format&amp;fit=facearea&amp;facepad=2&amp;w=256&amp;h=256&amp;q=80\" alt=\"\"></button></div><div x-show=\"userMenuOpen\" x-transition:enter=\"transition ease-out duration-200\" x-transition:enter-start=\"transform opacity-0 scale-95\" x-transition:enter-end=\"transform opacity-100 scale-100\" x-transition:leave=\"transition ease-in duration-75\" x-transition:leave-start=\"transform opacity-100 scale-100\" x-transition:leave-end=\"transform opacity-0 scale-95\" class=\"absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none\" role=\"menu\" aria-orientation=\"vertical\" aria-labelledby=\"user-menu-button\" tabindex=\"-1\" style=\"display: none;\"><a href=\"#\" class=\"block px-4 py-2 text-sm text-gray-700\" role=\"menuitem\" tabindex=\"-1\" id=\"user-menu-item-0\">Your Profile</a> <a href=\"#\" class=\"block px-4 py-2 text-sm text-gray-700\" role=\"menuitem\" tabindex=\"-1\" id=\"user-menu-item-1\">Settings</a><form action=\"/login\" method=\"post\"><button type=\"submit\" class=\"block px-4 py-2 text-sm text-gray-700\" role=\"menuitem\" tabindex=\"-1\" id=\"user-menu-item-2\">Sign out</button></form></div></div></div><div class=\"-mr-2 flex items-center sm:hidden\"><!-- Mobile menu button --><button type=\"button\" @click=\"mobileMenuOpen = !mobileMenuOpen\" class=\"relative inline-flex items-center justify-center rounded-md bg-white p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2\" :aria-expanded=\"mobileMenuOpen\"><span class=\"absolute -inset-0.5\"></span> <span class=\"sr-only\">Open main menu</span><!-- Menu open: \"hidden\", Menu closed: \"block\" --><svg x-show=\"!mobileMenuOpen\" class=\"block h-6 w-6\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" aria-hidden=\"true\" data-slot=\"icon\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5\"></path></svg><!-- Menu open: \"block\", Menu closed: \"hidden\" --><svg x-show=\"mobileMenuOpen\" class=\"h-6 w-6\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" aria-hidden=\"true\" data-slot=\"icon\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M6 18 18 6M6 6l12 12\"></path></svg></button></div></div></div><!-- Mobile menu --><div x-show=\"mobileMenuOpen\" class=\"sm:hidden\" id=\"mobile-menu\" style=\"display: none;\"><div class=\"space-y-1 pb-3 pt-2\"><a href=\"#\" class=\"block border-l-4 border-indigo-500 bg-indigo-50 py-2 pl-3 pr-4 text-base font-medium text-indigo-700\" aria-current=\"page\">Dashboard</a> <a href=\"#\" class=\"block border-l-4 border-transparent py-2 pl-3 pr-4 text-base font-medium text-gray-600 hover:border-gray-300 hover:bg-gray-50 hover:text-gray-800\">Team</a> <a href=\"#\" class=\"block border-l-4 border-transparent py-2 pl-3 pr-4 text-base font-medium text-gray-600 hover:border-gray-300 hover:bg-gray-50 hover:text-gray-800\">Projects</a> <a href=\"#\" class=\"block border-l-4 border-transparent py-2 pl-3 pr-4 text-base font-medium text-gray-600 hover:border-gray-300 hover:bg-gray-50 hover:text-gray-800\">Calendar</a></div><div class=\"border-t border-gray-200 pb-3 pt-4\"><div class=\"flex items-center px-4\"><div class=\"shrink-0\"><img class=\"h-10 w-10 rounded-full\" src=\"https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-1.2.1&amp;ixid=eyJhcHBfaWQiOjEyMDd9&amp;auto=format&amp;fit=facearea&amp;facepad=2&amp;w=256&amp;h=256&amp;q=80\" alt=\"\"></div><div class=\"ml-3\"><div class=\"text-base font-medium text-gray-800\">Tom Cook</div><div class=\"text-sm font-medium text-gray-500\">tom@example.com</div></div><button type=\"button\" class=\"relative ml-auto shrink-0 rounded-full bg-white p-1 text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2\"><span class=\"absolute -inset-1.5\"></span> <span class=\"sr-only\">View notifications</span> <svg class=\"h-6 w-6\" fill=\"none\" viewBox=\"0 0 24 24\" stroke-width=\"1.5\" stroke=\"currentColor\" aria-hidden=\"true\" data-slot=\"icon\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" d=\"M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0\"></path></svg></button></div><div class=\"mt-3 space-y-1\"><a href=\"#\" class=\"block px-4 py-2 text-base font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-800\">Your Profile</a> <a href=\"#\" class=\"block px-4 py-2 text-base font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-800\">Settings</a><form action=\"/login\" method=\"post\"><button type=\"button\" class=\"block px-4 py-2 text-base font-medium text-gray-500 hover:bg-gray-100 hover:text-gray-800\">Sign out</button></form></div></div></div></nav><div class=\"py-10\"><header><div class=\"mx-auto max-w-7xl px-4 sm:px-6 lg:px-8\"><h1 class=\"text-3xl font-bold tracking-tight text-gray-900\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(pageTitle)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/layout/app.templ`, Line: 123, Col: 78}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</h1></div></header><main><div class=\"mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("</div></main></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
