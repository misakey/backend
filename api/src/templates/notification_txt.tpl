{{.total}} nouveau(x) message(s) pour {{.displayName}} 

Pour accéder à mon application: https://app.misakey.com/?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=openApp

Les messages et documents envoyés dans les discussions sont protégés avec le chiffrement de bout-en-bout.

------------------------------------------------------------

Voici les détails de nouveau(x) message(s) (vous pouvez couper les notifications pour chaque espace sécurisé) :

{{ range $key, $value := .boxes }}
- {{ $value.Title }} : {{ $value.NewMessages }} message(s) non lu(s) (https://{{$.domain}}/boxes/{{$value.ID}}/details?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=openChannel)
{{ end }}

------------------------------------------------------------

Pour configurer la fréquence de réception des notifications : {{.accountBaseURL}}/notifications?utm_source=notification&utm_medium=email&utm_campaign=emailNotificationPreference&utm_content=notifsParamsFooter

------------------------------------------------------------

Cette adresse e-mail ne peut pas recevoir de réponse. Plus d’informations dans la section Aide de Misakey.
© Misakey SAS, 66 avenue des champs Elysée, 75008 Paris, France

 
