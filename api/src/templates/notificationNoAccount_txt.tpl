{{.total}} nouveau(x) message(s) pour {{.displayName}} 

Sans compte avec mot de passe, les liens d’invitations qui vous ont été partagés sont nécessaires pour accéder aux discussions protégées avec le chiffrement de bout-en-bout.
 
CRÉER UN COMPTE: https://app.misakey.com/?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=openApp

{{ range $key, $value := .boxes }}
- {{ $value.Title }} : {{ $value.NewMessages }} message(s) non lu(s)
{{ end }}
 
------------------------------------------------------------

Pour configurer la fréquence de réception des notifications : {{.accountBaseURL}}/notifications?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=notifsParamsFooter

------------------------------------------------------------
 
 This email address can't receive responses. If you want to know more, check Misakey help section. 
 © Misakey SAS, 66 avenue des champs Elysée, 75008 Paris, France
 
  
