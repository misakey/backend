module.exports = {
  title: 'Misakey documentation',
  tagline: 'Users deserve better privacy (docs under construction)',
  url: 'https://your-docusaurus-test-site.com',
  baseUrl: '/',
  noIndex: true,
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'throw',
  onDuplicateRoutes: 'throw',
  favicon: 'img/favicon.ico',
  organizationName: 'misakey',
  projectName: 'docs', // Usually your repo name.
  themeConfig: {
    respectPrefersColorScheme: true,
    announcementBar: {
      id: 'support_us', // Any value that will identify this message.
      content: 'This documentation is a complete work in progress. It\'s not usable, and only accessible for internal users.',
      backgroundColor: '#F9D2E1', // Defaults to `#fff`.
      textColor: '#091E42', // Defaults to `#000`.
      isCloseable: true, // Defaults to `true`.
    },
    navbar: {
      title: '',
      hideOnScroll: true,
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
