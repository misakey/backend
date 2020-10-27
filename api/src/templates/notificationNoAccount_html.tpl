<!DOCTYPE html>
<html>
    <head>
      <style type="text/css">
        img{line-height:100%}#outlook a{padding:0}.ReadMsgBody{width:100%}a,blockquote,body,li,p,table,td{-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%}table,td{mso-table-lspace:0;mso-table-rspace:0}img{-ms-interpolation-mode:bicubic;border:0;height:auto;outline:0;text-decoration:none}table{border-collapse:collapse!important}body{background-color:#d9d9d9}@font-face{font-family:Roboto;src:url(https://static.misakey.com/fonts/Roboto/Roboto-Medium.ttf);src:url(https://static.misakey.com/fonts/Roboto/Roboto-Medium.ttf) format('ttf'),url(https://static.misakey.com/fonts/Roboto/Roboto-Medium.woff) format('woff'),url(https://static.misakey.com/fonts/Roboto/Roboto-Medium.woff2) format('woff2');font-weight:400;font-style:normal}@font-face{font-family:Roboto;src:url(https://static.misakey.com/fonts/Roboto/Roboto-Bold.ttf);src:url(https://static.misakey.com/fonts/Roboto/Roboto-Bold.ttf) format('ttf'),url(https://static.misakey.com/fonts/Roboto/roboto-latin-700.woff) format('woff'),url(https://static.misakey.com/fonts/Roboto/roboto-latin-700.woff2) format('woff2');font-weight:600;font-style:normal}#bodyTable{width:600px;mso-table-lspace:0;mso-table-rspace:0;margin:0;margin-top:20px;padding:0;border:0;font-family:Roboto,sans-serif;border-collapse:collapse!important;background-color:#fff}a{color:#e32e72}hr{border-width:2px;border-style:solid;border-color:#eaeef3;border-bottom:0;margin-top:20px;margin-bottom:20px;margin-left:0;margin-right:0}#preheaderText{display:none!important;visibility:hidden;mso-hide:all;font-size:1px;color:#fff;line-height:1px;max-height:0;max-width:0;opacity:0;overflow:hidden}#logo{text-align:left;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%;margin-bottom:30px;margin-left:50px}#logo img{-ms-interpolation-mode:bicubic;border:0;height:auto;line-height:100%;outline:0;text-decoration:none;width:150px}#bodyTable td{-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%;mso-table-lspace:0;mso-table-rspace:0;margin:0;font-family:Roboto,sans-serif;text-align:center;padding:20px}.normal-weight{font-weight:400}.smalltext{font-size:.8em}#footer{text-align:center;color:#a9b3bc;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%}.bottom-border{border-bottom-style:solid;border-bottom-width:1px;border-bottom-color:#e32e72}@media only screen and (max-width:600px){a,blockquote,body,li,p,table,td{-webkit-text-size-adjust:none!important}body{min-width:100%!important}table{max-width:600px!important}}#channels-list{width:490px;margin-left:35px}#channels-list td{margin:0;padding:0;padding-top:2px;padding-bottom:2px;text-align:left}#channels-list td.channel-name{color:#696969;width:230px;text-align:left}#channels-list td.channel-unread{width:200px;padding-left:5px;padding-right:5px;color:#e32e72}#channels-list td.notif-off{text-align:right;font-size:.6em;width:50px}#channels-list td.notif-off a{color:#999}
      </style>
    </head>
    <body style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;background-color: #d9d9d9;min-width: 100%!important;">
      <center>
        <table cellpadding="0" cellspacing="0" id="bodyTable" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;width: 600px;margin: 0;margin-top: 20px;padding: 0;border: 0;font-family: Roboto,sans-serif;background-color: #fff;border-collapse: collapse!important;max-width: 600px!important;">
          <tr>
            <td id="preheaderText" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;visibility: hidden;mso-hide: all;font-size: 1px;color: #fff;line-height: 1px;max-height: 0;max-width: 0;opacity: 0;overflow: hidden;margin: 0;font-family: Roboto,sans-serif;text-align: center;padding: 20px;display: none!important;">
              {{.total}} nouveau(x) message(s) pour {{.displayName}}
            </td>
          </tr>
          <tr class="bottom-border" style="border-bottom-style: solid;border-bottom-width: 1px;border-bottom-color: #e32e72;">
            <td style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;margin: 0;font-family: Roboto,sans-serif;text-align: center;padding: 20px;">
              <p id="logo" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;text-align: left;margin-bottom: 30px;margin-left: 50px;">
                <a href="https://www.misakey.com?utm_source=notification&utm_medium=email&utm_campaign=emailConfirmationCode&utm_content=logo" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;color: #e32e72;">
                  <img src="https://static.misakey.com/img/MisakeyLogoTypo.png" alt="Misakey" style="line-height: 100%;-ms-interpolation-mode: bicubic;border: 0;height: auto;outline: 0;text-decoration: none;width: 150px;">
                </a>
              </p>

              <h2>{{.total}} <span class="normal-weight" style="font-weight: 400;">nouveau(x) message(s) pour</span></h2>
              <h2 class="normal-weight" style="font-weight: 400;">
                <!-- If we have an image -->
                {{ if .avatarURL }}
                <img src="{{.avatarURL}}" style="margin:0;line-height:100%;-ms-interpolation-mode:bicubic;border:0;outline:0;text-decoration:none;object-fit:cover;border-radius:50%;width:40px;height:40px;vertical-align:middle;">
                {{ else }}
                <!-- Else -->
                <span>
                  <!--[if mso]>
                    <v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" xmlns:w="urn:schemas-microsoft-com:office:word" style="height:40px;v-text-anchor:middle;width:40px;" arcsize="100%" strokecolor="#e32e72" fillcolor="#ffffff">
                      <w:anchorlock/>
                      <center style="color:#000000;font-family:sans-serif;font-size:18px;">{{.displayName}}</center>
                    </v:roundrect>
                  <![endif]-->
                  <span style="background-color:#ffffff;border:1px solid #e32e72;border-radius:40px;color:#000000;display:inline-block;font-family:sans-serif;font-size:18px;line-height:40px;text-align:center;text-decoration:none;width:40px;-webkit-text-size-adjust:none;mso-hide:all;">
                    {{.firstLetter}}
                  </span>
                </span>
                <!-- End -->
                {{ end }}
                <span>{{.displayName}}</span>
              </h2>

              <div><!--[if mso]>
                <v:roundrect xmlns:v="urn:schemas-microsoft-com:vml" xmlns:w="urn:schemas-microsoft-com:office:word" href="https://{{.domain}}/?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=openApp" style="height:40px;v-text-anchor:middle;width:350px;" arcsize="100%" stroke="f" fillcolor="#e32e72">
                  <w:anchorlock/>
                  <center>
                <![endif]-->
                    <a href="https://{{.domain}}/?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=openApp" style="background-color: #e32e72;border-radius: 40px;color: #ffffff;display: inline-block;font-family: sans-serif;font-size: 15px;line-height: 40px;text-align: center;text-decoration: none;width: 350px;-webkit-text-size-adjust: none;-ms-text-size-adjust: 100%;">
                      CRÉER UN COMPTE
                    </a>
                <!--[if mso]>
                  </center>
                </v:roundrect>
              <![endif]--></div>

              <p class="smalltext" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;font-size: .8em;">
                Sans compte avec mot de passe, les liens d’invitations qui vous ont été partagés sont nécessaires pour accéder aux discussions protégées avec le chiffrement de bout-en-bout.
              </p>
            </td>
          </tr>
          <tr class="bottom-border" style="border-bottom-style: solid;border-bottom-width: 1px;border-bottom-color: #e32e72;">
            <td style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;margin: 0;font-family: Roboto,sans-serif;text-align: center;padding: 20px;">
              <!-- Refacto with table, force channel name to 20chars max -->
              <table id="channels-list" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;width: 490px;margin-left: 35px;border-collapse: collapse!important;max-width: 600px!important;">
{{ range $key, $value := .boxes }}
                <tr>
                  <td class="channel-name" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;margin: 0;font-family: Roboto,sans-serif;text-align: left;padding: 0;padding-top: 2px;padding-bottom: 2px;color: #696969;width: 230px;">{{$value.Title}}&nbsp;:</td>
                  <td class="channel-unread" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;margin: 0;font-family: Roboto,sans-serif;text-align: left;padding: 0;padding-top: 2px;padding-bottom: 2px;width: 200px;padding-left: 5px;padding-right: 5px;color: #e32e72;">{{$value.NewMessages}} message(s) non lu(s)</td>
                </tr>
              </table>
            </td>
          </tr>
{{ end }}
          <tr>
            <td style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;mso-table-lspace: 0;mso-table-rspace: 0;margin: 0;font-family: Roboto,sans-serif;text-align: center;padding: 20px;">
              <p style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;">Pour configurer la fréquence de réception des notifications&nbsp;:</p>
              <a href="{{.accountBaseURL}}/notifications?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=notifsParamsFooter" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;color: #e32e72;">Paramètres de notifications</a>
              <hr style="border-width: 2px;border-style: solid;border-color: #eaeef3;border-bottom: 0;margin-top: 20px;margin-bottom: 20px;margin-left: 0;margin-right: 0;">
              <p id="footer" style="-webkit-text-size-adjust: 100%;-ms-text-size-adjust: 100%;text-align: center;color: #a9b3bc;">
                Cette adresse e-mail ne peut pas recevoir de réponse. Plus d’informations dans la section Aide de Misakey.
                <br> © Misakey SAS, 66 avenue des champs Elysée, 75008 Paris, France
              </p>
            </td>
          </tr>
        </table>
      </center>
    </body>
  </html>
  
