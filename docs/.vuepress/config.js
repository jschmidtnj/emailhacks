module.exports = {
  title: 'Mailpear Docs',
  description: 'Documentation for Mailpear',
  dest: 'emailhacks',
  head: [
    ['link', { rel: 'icon', href: `/logo.png` }],
    ['link', { rel: 'manifest', href: '/manifest.json' }],
    ['meta', { name: 'theme-color', content: '#3eaf7c' }],
    ['meta', { name: 'apple-mobile-web-app-capable', content: 'yes' }],
    ['meta', { name: 'apple-mobile-web-app-status-bar-style', content: 'black' }],
    ['link', { rel: 'apple-touch-icon', href: `/icons/apple-touch-icon-152x152.png` }],
    ['link', { rel: 'mask-icon', href: '/icons/safari-pinned-tab.svg', color: '#3eaf7c' }],
    ['meta', { name: 'msapplication-TileImage', content: '/icons/msapplication-icon-144x144.png' }],
    ['meta', { name: 'msapplication-TileColor', content: '#000000' }]
  ],
  serviceWorker: true,
  base: '/emailhacks/',
  cleanURL: true,
  themeConfig: {
    repo: 'jschmidtnj/emailhacks',
    editLinks: true,
    docsDir: 'docs',
    docsBranch: 'master',
    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'Prod', link: 'https://joshuaschmidt.tech' }
    ],
    displayAllHeaders: true,
    sidebar: 'auto'
  }
}
