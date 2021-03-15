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
      collapsed: false,
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
        'previous-documentation/home'
      ]
    },
  ]
};
