module.exports = {
  title: 'Misakey documentation',
  tagline: 'Users deserve better privacy (docs under construction)',
  url: 'https://docs.misakey.com',
  baseUrl: '/',
  noIndex: false,
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'throw',
  onDuplicateRoutes: 'throw',
  favicon: 'img/favicon.ico',
  organizationName: 'misakey',
  projectName: 'docs', // Usually your repo name.
  themeConfig: {
    respectPrefersColorScheme: true,
    algolia: {
      apiKey: 'caeec869399ae3eeb050e134fef631c9',
      indexName: 'misakey',

      // Optional: see doc section bellow
      contextualSearch: true,

      // Optional: Algolia search parameters
      searchParameters: {},

      //... other Algolia params
    },
    announcementBar: {
      id: 'beta', // Any value that will identify this message.
      content: 'This documentation is a work in progress. Feel free to <a href="mailto:love@misakey.com">email us</a> if you have any constructive feedback!',
      backgroundColor: '#F9D2E1', // Defaults to `#fff`.
      textColor: '#091E42', // Defaults to `#000`.
      isCloseable: true, // Defaults to `true`.
    },
    navbar: {
      title: '',
      hideOnScroll: false,
      logo: {
        alt: 'Misakey',
        src: 'https://static.misakey.com/img/MisakeyLogoTypo.svg',
      },
      items: [
        {
          to: 'docs/',
          activeBasePath: 'docs',
          label: 'Docs',
          position: 'left',
        },
        {
          href: 'https://github.com/misakey/backend',
          label: 'Source code',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Style Guide',
              to: 'docs/',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Stack Overflow',
              href: 'https://stackoverflow.com/questions/tagged/docusaurus',
            },
            {
              label: 'Discord',
              href: 'https://discordapp.com/invite/docusaurus',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com/docusaurus',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'Website',
              href: 'https://www.misakey.com',
            },
            {
              label: 'Blog',
              href: 'https://blog.misakey.com'
            },
            {
              label: 'GitHub',
              href: 'https://github.com/misakey/backend',
            },
          ],
        },
      ],
      copyright: `Content published on Creative Commmon  CC BY-SA 4.0 License - Misakey. Built with Docusaurus.`,
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl:
            'https://gitlab.com/misakey/docs/edit/master/website/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            'https://gitlab.com/misakey/docs/edit/master/website/blog/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
