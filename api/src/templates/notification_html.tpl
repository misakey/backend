<!DOCTYPE html>
<html>
    <head>
      <style type="text/css">
        @media only screen and (max-width:600px){a,blockquote,body,li,p,table,td{-webkit-text-size-adjust:none}body{min-width:100%}table{max-width:600px}}
      </style>
    </head>
    <body style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; background-color:#d9d9d9" bgcolor="#d9d9d9">
      <center>
        <table cellpadding="0" cellspacing="0" id="bodyTable" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; mso-table-lspace:0; mso-table-rspace:0; background-color:#FFF; border:0; font-family:Roboto, sans-serif; margin:0; margin-top:20px; padding:0; width:600px; border-collapse:collapse" bgcolor="#FFFFFF" width="600">
          <tr>
            <td id="preheaderText" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; mso-table-lspace:0; mso-table-rspace:0; color:#fff; font-size:1px; line-height:1px; max-height:0; max-width:0; mso-hide:all; opacity:0; overflow:hidden; visibility:hidden; font-family:Roboto, sans-serif; margin:0; padding:20px; text-align:center; display:none" align="center">
                {{.total}} nouveau(x) message(s) pour {{.displayName}}
            </td>
          </tr>
          <tr style="border-bottom-color:#e32e72; border-bottom-style:solid; border-bottom-width:1px">
            <td style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; mso-table-lspace:0; mso-table-rspace:0; font-family:Roboto, sans-serif; margin:0; padding:20px; text-align:center" align="center">
              <div id="logo" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; margin-bottom:30px; margin-left:50px; text-align:left" align="left">
                <a href="https://www.misakey.com?utm_source=notification&amp;utm_medium=email&amp;utm_campaign=emailConfirmationCode&amp;utm_content=logo" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; color:#e32e72">
                  <img src="https://static.misakey.com/img/MisakeyLogoTypo.png" alt="Misakey" style="line-height:100%; -ms-interpolation-mode:bicubic; border:0; height:auto; outline:none; text-decoration:none; width:150px" height="auto" width="150">
                </a>
              </div>

              <h2>{{.total}} <span style="font-weight:normal">nouveau(x) message(s) pour</span></h2>
              <h2 style="font-weight:normal">
                <!-- If we have an image -->
                {{ if .avatarURL }}
                <img src="{{.avatarURL}}" width="40" height="40" style="line-height:100%; -ms-interpolation-mode:bicubic; border:0; height:40px; outline:0; text-decoration:none; border-radius:40px; margin:0; width:40px" alt="{{.firstLetter}}">
                {{ else }}
                <!-- Else -->
                <span>
                  <!--[if mso]>
                    <v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" xmlns:w="urn:schemas-microsoft-com:office:word" style="height:40px;v-text-anchor:middle;width:40px;" arcsize="100%" strokecolor="#e32e72" fillcolor="#ffffff">
                      <w:anchorlock/>
                      <center style="color:#000000;font-family:sans-serif;font-size:18px;">{{.firstLetter}}</center>
                    </v:roundrect>
                  <![endif]-->
                  <span style="background-color:#ffffff;border:1px solid #e32e72;border-radius:40px;color:#000000;display:inline-block;font-family:sans-serif;font-size:18px;line-height:40px; height:40px;text-align:center;text-decoration:none;width:40px;-webkit-text-size-adjust:none;mso-hide:all;">
                    {{.firstLetter}}
                  </span>
                </span>
                <!-- End -->
                {{ end }}
                <span>{{.displayName}}</span>
              </h2>

              <div>
<!--[if mso]>
                <v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" xmlns:w="urn:schemas-microsoft-com:office:word" href="https://{{.domain}}/?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=openApp" style="height:40px;v-text-anchor:middle;width:350px;" arcsize="100%" stroke="f" fillcolor="#e32e72">
                  <w:anchorlock/>
                  <center>
                <![endif]-->
                    <a href="https://{{.domain}}/?utm_source=notification&amp;utm_medium=email&amp;utm_campaign=emailNotificationPreference&amp;utm_content=openApp" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:none; color:#fff; background-color:#e32e72; border-radius:40px; display:inline-block; font-family:sans-serif; font-size:15px; height:40yypx; line-height:40px; text-align:center; text-decoration:none; width:350px" bgcolor="#e32e72" height="40yy" align="center" width="350">OUVRIR MON APPLICATION</a>
                <!--[if mso]>
                  </center>
                </v:roundrect>
              <![endif]-->
</div>

              <p style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; font-size:0.8em">
                Les messages et documents envoyés dans les discussions sont protégés avec le chiffrement de bout-en-bout.
              </p>
            </td>
          </tr>
          <tr style="border-bottom-color:#e32e72; border-bottom-style:solid; border-bottom-width:1px">
            <td style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; mso-table-lspace:0; mso-table-rspace:0; font-family:Roboto, sans-serif; margin:0; padding:20px; text-align:center" align="center">
              <!-- Refacto with table, force channel name to 20chars max -->
              <table id="channels-list" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; mso-table-lspace:0; mso-table-rspace:0; margin-left:35px; width:490px; border-collapse:collapse" width="490">
{{ range $key, $value := .boxes }}
                <tr>
                  <td class="channel-name" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0pt;mso-table-rspace: 0pt;margin: 0;font-family: Roboto, sans-serif;text-align: left;padding: 0;padding-top: 2px;padding-bottom: 2px;color: #696969;width: 230px;">{{ $value.Title }}&nbsp;:</td>
                  <td class="channel-unread" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0pt;mso-table-rspace: 0pt;margin: 0;font-family: Roboto, sans-serif;text-align: left;padding: 0;padding-top: 2px;padding-bottom: 2px;width: 200px;padding-left: 5px;padding-right: 5px;"><a href="https://{{$.domain}}/boxes/{{$value.ID}}?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=openChannel" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;color: #e32e72;">{{ $value.NewMessages}} message(s) non lu(s)</a></td>
                  <td class="notif-off" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0pt;mso-table-rspace: 0pt;margin: 0;font-family: Roboto, sans-serif;text-align: right;padding: 0;padding-top: 2px;padding-bottom: 2px;font-size: 0.6em;width: 50px;"><a href="https://{{$.domain}}/boxes/{{$value.ID}}/details?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=notifOff" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;color: #999999;">(Notif off)</a></td>
                </tr>
{{ end }}
              </table>
            </td>
          </tr>
          <tr>
            <td style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; mso-table-lspace:0; mso-table-rspace:0; font-family:Roboto, sans-serif; margin:0; padding:20px; text-align:center" align="center">
              <p style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%">Pour configurer la fréquence de réception des notifications :</p>
              <a href="{{.accountBaseURL}}/notifications?utm_source=notification&amp;utm_medium=email&amp;utm_campaign=emailNotificationPreference&amp;utm_content=notifsParamsFooter" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; color:#e32e72">Paramètres de notifications</a>
              <hr style="border-bottom:0; border-color:#EAEEF3; border-style:solid; border-width:2px; margin-bottom:20px; margin-left:0; margin-right:0; margin-top:20px">
              <p id="footer" style="-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%; color:#A9B3BC; text-align:center" align="center">
                Cette adresse e-mail ne peut pas recevoir de réponse. Plus d’informations dans la section Aide de Misakey.
                <br> © Misakey SAS, 66 avenue des champs Elysée, 75008 Paris, France
              </p>
            </td>
          </tr>
        </table>
      </center>
    </body>
  </html>