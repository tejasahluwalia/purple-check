// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.857
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

var validateInstagramUsername = templ.NewOnceHandle()

func Search() templ.Component {
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
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<form action=\"/search\" method=\"post\" class=\"block\" id=\"search-form\"><label for=\"search-term\" class=\"block font-medium\">Search for a user to see reviews</label><div class=\"flex flex-col w-full space-y-4 mt-1 relative\"><div class=\"relative\"><input placeholder=\"@username\" autocapitalize=\"none\" class=\"block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-purple-600 sm:text-sm/6\" type=\"text\" name=\"search-term\" id=\"search-term\" required><p id=\"error-message\" class=\"hidden text-red-500 text-sm mt-1\"></p></div><input class=\"rounded-md bg-purple-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-sm hover:not-disabled:bg-purple-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-purple-600 transition-colors\" type=\"submit\" id=\"search-button\" value=\"Search\" aria-label=\"Search for an Instagram user.\" disabled></div></form><script>\n        const form = document.getElementById('search-form');\n        const input = document.getElementById('search-term');\n        const button = document.getElementById('search-button');\n        const errorMessage = document.getElementById('error-message');\n\n        function validateInstagramUsername(username) {\n            // Remove @ if present\n            username = username.replace('@', '');\n\n            // Instagram username rules:\n            // 1. 1-30 characters\n            // 2. Can contain letters, numbers, periods, and underscores\n            // 3. Can't start or end with a period\n            // 4. Can't have consecutive periods\n            const allowedCharacters = \"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._\";\n            const hasConsecutivePeriods = /\\.{2,}/.test(username);\n\n            return username.length >= 1 &&\n                   username.length <= 30 &&\n                   !hasConsecutivePeriods &&\n                   [...username].every(char => allowedCharacters.includes(char));\n        }\n\n        input.addEventListener('input', function(e) {\n            const username = input.value.trim();\n\n            if (!validateInstagramUsername(username)) {\n                errorMessage.textContent = 'Please enter a valid Instagram username (1-30 characters, letters, numbers, periods, or underscores only).';\n                errorMessage.classList.remove('hidden');\n                input.classList.remove('ring-slate-300');\n                input.classList.add('ring-red-500');\n                button.disabled = true;\n            } else {\n                errorMessage.classList.add('hidden');\n                input.classList.remove('ring-red-500');\n                input.classList.add('ring-slate-300');\n                button.disabled = false;\n            }\n        });\n\n        button.addEventListener('click', function(e) {\n\t\t\t\t\t\tform.submit();\n\t\t\t\t});\n    </script>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
