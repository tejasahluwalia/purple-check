// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.833
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"log"
	"purple-check/internal/database"
	"purple-check/internal/models"
)

func getFeedbackList(username string, role string) []models.Feedback {
	db, closer := database.GetDB()
	defer closer()

	var feedbackList []models.Feedback

	stmt, err := db.Prepare("SELECT id, giver, receiver, rating, created_at FROM feedback WHERE " + role + " = ? ORDER BY created_at DESC")
	if err != nil {
		log.Println(err)
	}

	rows, err := stmt.Query(username)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var feedback models.Feedback

		err = rows.Scan(&feedback.ID, &feedback.Giver, &feedback.Receiver, &feedback.Rating, &feedback.CreatedAt)
		if err != nil {
			log.Println(err)
		}

		feedbackList = append(feedbackList, feedback)
	}

	return feedbackList
}

var dateFormatter = templ.NewOnceHandle()

func FeedbackList(p string) templ.Component {
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

		feedbackList := getFeedbackList(p, "receiver")
		if len(feedbackList) == 0 {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<p class=\"text-slate-500\">No feedback yet.</p>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		} else {
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<ul class=\"space-y-8 divide-y text-slate-500 divide-slate-200 border-b border-t border-slate-200 pb-10\" hx-confirm=\"Are you sure?\" hx-target=\"closest .feedback\" hx-swap=\"outerHTML\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			for _, feedback := range feedbackList {
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "<li class=\"feedback pt-4 grid\"><div class=\"mt-2 mb-4 text-slate-500\"><p>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var2 string
				templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(feedback.Rating)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/components/feedback-list.templ`, Line: 53, Col: 24}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "</p></div><div class=\"flex space-x-2 items-center\"><h3 class=\"font-medium text-slate-900\"><a hx-boost=\"false\" href=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var3 templ.SafeURL = templ.URL("/profile/" + feedback.Giver)
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(string(templ_7745c5c3_Var3)))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var4 string
				templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs(feedback.Giver)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/components/feedback-list.templ`, Line: 58, Col: 92}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "</a></h3><span>.</span><p class=\"text-slate-500\"><time datetime=\"")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var5 string
				templ_7745c5c3_Var5, templ_7745c5c3_Err = templ.JoinStringErrs(feedback.CreatedAt)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/components/feedback-list.templ`, Line: 64, Col: 42}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var5))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 7, "\" class=\"created-at-datetime\">")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				var templ_7745c5c3_Var6 string
				templ_7745c5c3_Var6, templ_7745c5c3_Err = templ.JoinStringErrs(feedback.CreatedAt)
				if templ_7745c5c3_Err != nil {
					return templ.Error{Err: templ_7745c5c3_Err, FileName: `internal/components/feedback-list.templ`, Line: 64, Col: 93}
				}
				_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var6))
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
				templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 8, "</time></p></div></li>")
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 9, "</ul>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
		}
		templ_7745c5c3_Var7 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 10, "<script>\n        document.querySelectorAll('.created-at-datetime').forEach((datetime) => {\n            const createdAt = datetime.getAttribute('datetime');\n            const date = new Date(createdAt);\n\n            const since = Date.now() - date.getTime();\n            if (since < 1000 * 60) {\n                datetime.innerHTML = 'Just now';\n                return;\n            }\n            if (since < 1000 * 60 * 60) {\n                const minutes = Math.floor(since / (1000 * 60));\n                datetime.innerHTML = `${minutes} minute${minutes > 1 ? 's' : ''} ago`;\n                return;\n            }\n            if (since < 1000 * 60 * 60 * 24) {\n                const hours = Math.floor(since / (1000 * 60 * 60));\n                datetime.innerHTML = `${hours} hour${hours > 1 ? 's' : ''} ago`;\n                return;\n            }\n\n            datetime.innerHTML = date.toLocaleDateString(undefined, {\n                month: 'short',\n                day: 'numeric',\n                year: 'numeric',\n            });\n        });\n    \t</script>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = dateFormatter.Once().Render(templ.WithChildren(ctx, templ_7745c5c3_Var7), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
