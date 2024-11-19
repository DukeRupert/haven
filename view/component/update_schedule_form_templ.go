// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.793
package component

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/DukeRupert/haven/db"
	"github.com/DukeRupert/haven/types"
)

func UpdateScheduleForm(route types.RouteContext, schedule db.Schedule) templ.Component {
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
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div id=\"update-schedule-form\" hx-target=\"this\" hx-target-error=\"#global-alert\" hx-swap=\"outerHTML\" class=\"relative lg:col-span-2\"><div class=\"h-full overflow-hidden rounded-lg bg-white shadow\"><div class=\"px-6 py-8\"><div class=\"flex items-center justify-between\"><h3 class=\"text-lg font-medium text-gray-900\">Update Schedule</h3></div><div class=\"mt-6\"><form hx-put=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(route.BuildURL("/schedule"))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/update_schedule_form.templ`, Line: 17, Col: 42}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" hx-target=\"#update-schedule-form\" hx-swap=\"outerHTML\" hx-target-error=\"#global-alert\"><div class=\"space-y-6\"><div class=\"grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2\"><div><label for=\"first_weekday\" class=\"block text-sm font-medium text-gray-700\">First Weekday</label> <select id=\"first_weekday\" name=\"first_weekday\" class=\"mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm\" required><option value=\"0\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.FirstWeekday == 0 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Sunday</option> <option value=\"1\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.FirstWeekday == 1 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Monday</option> <option value=\"2\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.FirstWeekday == 2 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Tuesday</option> <option value=\"3\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.FirstWeekday == 3 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Wednesday</option> <option value=\"4\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.FirstWeekday == 4 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Thursday</option> <option value=\"5\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.FirstWeekday == 5 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Friday</option> <option value=\"6\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.FirstWeekday == 6 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Saturday</option></select></div><div><label for=\"second_weekday\" class=\"block text-sm font-medium text-gray-700\">Second Weekday</label> <select id=\"second_weekday\" name=\"second_weekday\" class=\"mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm\" required><option value=\"0\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.SecondWeekday == 0 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Sunday</option> <option value=\"1\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.SecondWeekday == 1 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Monday</option> <option value=\"2\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.SecondWeekday == 2 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Tuesday</option> <option value=\"3\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.SecondWeekday == 3 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Wednesday</option> <option value=\"4\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.SecondWeekday == 4 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Thursday</option> <option value=\"5\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.SecondWeekday == 5 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Friday</option> <option value=\"6\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		if schedule.SecondWeekday == 6 {
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(" selected")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(">Saturday</option></select></div><div class=\"sm:col-span-2\"><label for=\"start_date\" class=\"block text-sm font-medium text-gray-700\">Start Date</label> <input type=\"date\" id=\"start_date\" name=\"start_date\" value=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var3 string
		templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs(schedule.StartDate.Format("2006-01-02"))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/update_schedule_form.templ`, Line: 64, Col: 57}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" class=\"mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm\" required></div></div><div class=\"flex justify-end space-x-3\"><button type=\"button\" hx-get=\"")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var4 string
		templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(route.BuildURL("/schedule"))
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `view/component/update_schedule_form.templ`, Line: 73, Col: 45}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("\" hx-target=\"#update-schedule-form\" hx-swap=\"outerHTML\" class=\"rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2\">Cancel</button> <button type=\"submit\" class=\"inline-flex justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2\">Update</button></div></div></form></div></div></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate