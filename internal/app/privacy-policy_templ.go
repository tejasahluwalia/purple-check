// Code generated by templ - DO NOT EDIT.

// templ: version: v0.2.771
package app

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func PrivacyPolicy() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
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
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString("<div class=\"flex w-full justify-center\"><section class=\"mx-auto block px-4 py-8 prose\"><h1>Privacy Policy</h1><p>We value your privacy. With this policy, we set out your privacy rights and how we collect, use, disclose, transfer and store your personal data.</p><h2>Terms we use in this policy</h2><p>When we say \"Purple Check\", “we”, “our”, or “us”, we mean Tujux Labs (OPC) Private Limited, which is the entity responsible for processing your personal data.</p><h2>We&apos;re an open platform</h2><p>Because we&apos;re an open platform, when you leave feedback, the feedback and your profile will be visible to anyone who visits our platform. By clicking on your profile, such visitors can see all the feedback you&apos;ve left.</p><h2>Personal data we collect</h2><p>Personal data is any information that relates to an identifiable individual. When you connect your social media account, we may collect and process the following personal data:</p><ul><li><span class=\"font-bold\">Connected social networks: </span> Before your request to connect with a social network profile is carried out, you&apos;ll be told which information we will collect from that social network. You can disconnect your social network profile at any time. We will then remove your social network unique ID from our database.</li><li><span class=\"font-bold\">Information you provide: </span> When you contact us, we may ask for your name, email address, and other personal information.</li><li><span class=\"font-bold\">Usage and profiling information: </span> Your search history, how you've interacted with our platform, or emails we send to you, including time you spend on our site, features or functions you&apos;ve accessed, marketing emails you&apos;ve opened and links you&apos;ve clicked.</li><li><span class=\"font-bold\">Information about feedback and ratings, including:</span><br><ul><li>Which page you left feedback on.</li><li>Your feedback content and star rating.</li></ul></li></ul><h2>Why and how we use your personal data</h2><p>We may use your personal data to:<br><ul><li>Provide you with our services.</li><li>Verify the authenticity of your feedback.</li><li>Improve our services.</li><li>Respond to your questions and provide customer service.</li><li>Engage in various internal business purposes, such as data analysis, audits, fraud monitoring and prevention, developing new products and services.</li><li>Exercise or comply with our own legal rights or obligations in connection with legal claims, or for compliance, regulatory and auditing purposes, where necessary.</li></ul></p><h2>Who may access your personal data?</h2><p>We share your feedback on our platform so that others can read about your experience. When you leave feedback on a page, other users of our platform will be able to see your username.</p><h2>How do we keep your personal data secure?</h2><p>Keeping your personal data secure is our highest priority. We use various organizational, technical and administrative measures to protect your personal data within our organization and we regularly audit our system for vulnerabilities. However, since the internet is not a completely secure environment, we can&apos;t ensure or warrant the security of the information you transmit to us. Emails sent via the platform may not be encrypted, and we therefore advise you not to include any confidential information in your emails to us.</p><h2>Do we use cookies?</h2><p>We do not use any third party cookies or tracking codes or tracking pixels.</p><h2>Your rights</h2><p>You can request to delete all your information at any time.<br><a href=\"http://www.purple-check.com/delete\">Delete my data</a></p></section></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return templ_7745c5c3_Err
	})
}

var _ = templruntime.GeneratedTemplate
