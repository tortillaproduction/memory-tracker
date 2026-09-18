package notify_overdue_sites

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

type emailSiteView struct {
	Name        string
	URL         string
	StatusLabel string
}

type emailData struct {
	UserName  string
	Sites     []emailSiteView
	SiteCount int
}

const emailHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<title>Memory Tracker</title>
</head>
<body style="margin:0; padding:0; background-color:#EEEBF6;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#EEEBF6;">
  <tr>
    <td align="center" style="padding: 40px 16px;">
      <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="width:600px; max-width:600px;">

        <!-- card -->
        <tr>
          <td style="background-color:#FFFFFF; border-radius:16px; overflow:hidden;">

            <!-- gradient header / wordmark -->
            <!--[if mso]>
            <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0"><tr><td bgcolor="#3B1160" style="padding:56px 32px;">
            <![endif]-->
            <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#3B1160" style="background-color:#3B1160; background-image:linear-gradient(160deg,#1B0F30 0%,#3B1160 55%,#6D28D9 100%);">
              <tr>
                <td align="center" style="padding:56px 32px 50px 32px;">
                  <div style="font-family: Georgia, 'Times New Roman', serif; font-size:38px; line-height:1.2; font-weight:bold; color:#FFFFFF; letter-spacing:0.5px; text-shadow: 0 0 2px #FFFFFF, 0 0 4px #FFFFFF, 0 0 8px #C4B5FD, 0 0 16px #A78BFA, 0 0 28px #7C3AED, 0 0 46px #6D28D9, 0 0 4px #F0ABFC;">
                    Memory&nbsp;Tracker
                  </div>
                  <div style="padding-top:10px; font-family: Arial, Helvetica, sans-serif; font-size:13px; letter-spacing:1.5px; text-transform:uppercase; color:#D8CCF0;">
                    before forget
                  </div>
                </td>
              </tr>
            </table>
            <!--[if mso]>
            </td></tr></table>
            <![endif]-->

            <!-- body -->
            <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
              <tr>
                <td style="padding:36px 40px 40px 40px;">

                  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                    <tr>
                      <td style="font-family: Arial, Helvetica, sans-serif; font-size:16px; font-weight:bold; color:#6D28D9; padding-bottom:12px;">
                        Hi {{.UserName}},
                      </td>
                    </tr>
                    <tr>
                      <td style="font-family: Arial, Helvetica, sans-serif; font-size:14px; line-height:1.75; color:#635C77;">
                        The sites below are past their check-in interval. Open a site using the button and your visit will be recorded automatically.
                      </td>
                    </tr>
                  </table>

                  <!-- overview box -->
                  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin-top:24px;">
                    <tr>
                      <td style="background-color:#F6F4FB; border-radius:10px; padding:2px 22px 18px 22px;">
                        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                          <tr>
                            <td style="padding-top:18px; padding-bottom:6px; font-family: Arial, Helvetica, sans-serif; font-size:13px; font-weight:bold; color:#241C33;">
                              Overdue sites ({{.SiteCount}})
                            </td>
                          </tr>
                          {{range .Sites}}
                          <tr>
                            <td style="padding:8px 0; font-family: Arial, Helvetica, sans-serif; font-size:13.5px; color:#3D3653; border-top:1px solid #E1DBF0;">
                              <span style="color:#A78BFA; font-weight:bold;">&#9679;</span>&nbsp; {{.Name}}&nbsp;&nbsp;
                              <span style="color:#635C77;">{{.StatusLabel}}</span>
                            </td>
                          </tr>
                          {{end}}
                        </table>
                      </td>
                    </tr>
                  </table>

                  <!-- CTA buttons, one per site, labelled with the site name -->
                  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin-top:28px;">
                    {{range .Sites}}
                    <tr>
                      <td align="center" style="padding-bottom:14px;">
                        <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                          <tr>
                            <td bgcolor="#7C3AED" style="background-color:#7C3AED; background-image:linear-gradient(135deg,#A78BFA 0%,#7C3AED 100%); border-radius:24px;">
                              <a href="{{.URL}}" target="_blank" style="display:inline-block; padding:13px 40px; font-family: Arial, Helvetica, sans-serif; font-size:14px; font-weight:bold; color:#FFFFFF; text-decoration:none; border-radius:24px;">{{.Name}}</a>
                            </td>
                          </tr>
                        </table>
                      </td>
                    </tr>
                    {{end}}
                  </table>

                  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin-top:16px;">
                    <tr>
                      <td style="font-family: Arial, Helvetica, sans-serif; font-size:13px; line-height:1.7; color:#635C77;">
                        Just reply to this email if you have any questions.
                      </td>
                    </tr>
                  </table>

                </td>
              </tr>
            </table>

          </td>
        </tr>

        <!-- footer -->
        <tr>
          <td align="center" style="padding-top:24px; font-family: Arial, Helvetica, sans-serif; font-size:11.5px; line-height:1.8; color:#8A82A3;">
            <a href="{{.PrivacyURL}}" target="_blank" style="color:#8A82A3; text-decoration:underline;">Privacy Policy</a>
            &nbsp;&middot;&nbsp;
            <a href="{{.TermsURL}}" target="_blank" style="color:#8A82A3; text-decoration:underline;">Terms of Service</a>
            <br>
            &#169; {{.Year}} Memory Tracker. All rights reserved.
          </td>
        </tr>

      </table>
    </td>
  </tr>
</table>
</body>
</html>`

func buildEmailHTML(userName string, sites []*SiteRow, now time.Time, frontendURL string) (string, error) {
	data := struct {
		UserName   string
		Sites      []emailSiteView
		SiteCount  int
		Year       int
		PrivacyURL string
		TermsURL   string
	}{
		UserName:   userName,
		SiteCount:  len(sites),
		Year:       now.Year(),
		PrivacyURL: frontendURL + "/privacy",
		TermsURL:   frontendURL + "/terms",
	}

	for _, s := range sites {
		var status string
		if s.LastCheckedAt != nil {
			hours := now.Sub(*s.LastCheckedAt).Hours()
			status = fmt.Sprintf("last checked %.0fh ago - every %dh", hours, s.IntervalHours)
		} else {
			status = fmt.Sprintf("not checked yet - every %dh", s.IntervalHours)
		}
		data.Sites = append(data.Sites, emailSiteView{
			Name: s.SiteName,
			// s.CheckinURLはワンタイムトークン付きの/go/{siteId}リンク。
			// s.SiteURLを直接使うと経由せずに開けてしまいチェックインが記録されない。
			URL:         s.CheckinURL,
			StatusLabel: status,
		})
	}

	tmpl, err := template.New("overdue_sites_email").Parse(emailHTMLTemplate)
	if err != nil {
		return "", fmt.Errorf("parse email template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render email template: %w", err)
	}

	return buf.String(), nil
}
