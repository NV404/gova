import type { ReactNode } from 'react';
import type { Metadata } from 'next';
import { RootProvider } from 'fumadocs-ui/provider';
import 'fumadocs-ui/style.css';
import './global.css';

export const metadata: Metadata = {
  metadataBase: new URL('https://gova.dev'),
  title: {
    default: 'Gova · The declarative GUI framework for Go',
    template: '%s · Gova',
  },
  description:
    'The declarative GUI framework for Go. Build native desktop apps for macOS, Windows, and Linux with typed components, reactive state, real platform dialogs, and one static binary.',
  keywords: [
    'Go',
    'Golang',
    'GUI',
    'desktop',
    'declarative',
    'Fyne',
    'cross-platform',
  ],
  icons: {
    icon: [
      { url: '/gova-logo.svg', type: 'image/svg+xml' },
    ],
    shortcut: '/gova-logo.svg',
    apple: '/gova-logo.svg',
  },
  openGraph: {
    title: 'Gova · The declarative GUI framework for Go',
    description:
      'Build native desktop apps in Go with typed components, reactive state, and real platform dialogs.',
    type: 'website',
    url: 'https://gova.dev',
    siteName: 'Gova',
  },
};

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      className="dark"
      style={{ colorScheme: 'dark' }}
      suppressHydrationWarning
    >
      <body className="gova-grain">
        <RootProvider theme={{ enabled: false }}>
          {children}
        </RootProvider>
      </body>
    </html>
  );
}
