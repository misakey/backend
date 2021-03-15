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
      contextualSearch: false,
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
          title: 'Main',
          items: [
            {
              label: 'Website',
              href: 'https://www.misakey.com',
            },
            {
              label: 'Demo',
              href: 'https://app.misakey.com',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/misakey',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Twitter',
              href: 'https://twitter.com/gomisakey',
            },
            {
              label: 'Linkedin',
              href: 'https://www.linkedin.com/company/gomisakey',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'French Blog',
              href: 'https://blog.misakey.com'
            },
            {
              label: 'Encryption White Paper',
              href: 'https://about.misakey.com/cryptography/white-paper.html'
            },
            {
              label: 'About',
              href: 'https://about.misakey.com/#/fr/',
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
