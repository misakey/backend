module.exports = {
  mainSidebar: [
    {
      type: 'doc',
      id: 'home'
    },
    {
      type: 'category',
      label: 'Concepts',
      items: [
        'concepts/organizations',
        'concepts/organizations',
        'concepts/datatags',
        'concepts/boxes',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/overview',
        'guides/your-org',
        'guides/store-data-for-your-org',
        'guides/self-hosting-misakey',
      ],
      collapsed: false,
    },
    {
      type: 'category',
      label: 'References',
      items: [
        'references/overview',
        'references/boxes',
        'references/datatags',
        'references/authorizations',
        'references/identities',
        'references/aes-rsa-encryption',
        'references/errors-format',
      ],
    },
    {
      type: 'category',
      label: 'Integrations',
      items: [
        'integrations/overview',
        'integrations/sdk-integrations',
        'integrations/cms-integrations',
      ],
    },
    {
      type: 'category',
      label: 'Previous documentation',
      items: [
        {
          type: 'doc',
          id: 'old-doc/home'
        },
        {
          type: 'category',
          label: 'Concepts',
          items: [
            "old-doc/concepts/authzn",
            "old-doc/concepts/box-events",
            "old-doc/concepts/identity-public-keys",
            "old-doc/concepts/quota",
            "old-doc/concepts/realtime",
            "old-doc/concepts/server-relief"
          ]
        },
        {
          type: 'category',
          label: 'Endpoints',
          items: [
            "old-doc/endpoints/accounts",
            "old-doc/endpoints/auth_flow",
            "old-doc/endpoints/backup-archives",
            "old-doc/endpoints/box_enc_files",
            "old-doc/endpoints/box_events",
            "old-doc/endpoints/box_key_shares",
            "old-doc/endpoints/box_saved_files",
            "old-doc/endpoints/box_users",
            "old-doc/endpoints/boxes",
            "old-doc/endpoints/coupons",
            "old-doc/endpoints/crypto_actions",
            "old-doc/endpoints/datatag",
            "old-doc/endpoints/generic",
            "old-doc/endpoints/identities",
            "old-doc/endpoints/organizations",
            "old-doc/endpoints/quota",
            "old-doc/endpoints/realtime",
            "old-doc/endpoints/root_key_shares",
            "old-doc/endpoints/secret_storage",
            "old-doc/endpoints/totp",
            "old-doc/endpoints/webauthn"
          ]
        }
      ]
    },
  ]
};


