import Link from 'next/link';
import type { CSSProperties, ReactNode } from 'react';

export default function HomePage() {
  return (
    <main style={{ position: 'relative', zIndex: 2 }}>
      <Hero />
      <Metrics />
      <Features />
      <Platforms />
      <Toolchain />
      <Quickstart />
      <CTA />
      <Footer />
    </main>
  );
}

/* ---------- Hero ---------- */

function Hero() {
  return (
    <section
      style={{
        position: 'relative',
        overflow: 'hidden',
        borderBottom: '1px solid var(--gova-line)',
      }}
    >
      <div
        aria-hidden
        className="gova-grid gova-grid-fade"
        style={{ position: 'absolute', inset: 0, zIndex: 0 }}
      />
      <div
        className="gova-hero-grid"
        style={{
          position: 'relative',
          zIndex: 1,
          maxWidth: 1200,
          margin: '0 auto',
          padding:
            'clamp(3.5rem, 7vw, 5.5rem) clamp(1.25rem, 4vw, 2.5rem) clamp(3rem, 6vw, 4.5rem)',
          display: 'grid',
          gridTemplateColumns: 'minmax(0, 1.1fr) minmax(0, 0.9fr)',
          gap: 'clamp(2rem, 4vw, 3.5rem)',
          alignItems: 'center',
        }}
      >
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: '1.4rem',
          }}
        >
          <h1
            className="gova-rise"
            style={{
              animationDelay: '60ms',
              margin: 0,
              fontFamily: 'var(--font-sans)',
              fontWeight: 500,
              fontSize: 'clamp(2.4rem, 5.6vw, 4.4rem)',
              lineHeight: 1.02,
              letterSpacing: '-0.038em',
              color: 'var(--gova-cream)',
            }}
          >
            The declarative GUI framework for Go.
          </h1>

          <p
            className="gova-rise"
            style={{
              animationDelay: '120ms',
              margin: 0,
              maxWidth: 560,
              fontSize: 'clamp(1rem, 1.05vw, 1.08rem)',
              lineHeight: 1.62,
              color: 'var(--gova-cream-dim)',
            }}
          >
            Build native desktop apps for macOS, Windows, and Linux from a single
            Go codebase. Typed components, reactive state, real platform dialogs,
            and one static binary.
          </p>

          <InstallCommand />

          <div
            className="gova-rise"
            style={{
              animationDelay: '240ms',
              display: 'flex',
              gap: '0.65rem',
              flexWrap: 'wrap',
            }}
          >
            <Link
              href="/docs/getting-started/installation"
              className="gova-btn gova-btn--primary"
            >
              Get started <Arrow />
            </Link>
            <Link href="/docs" className="gova-btn gova-btn--ghost">
              Documentation
            </Link>
            <Link
              href="https://github.com/nv404/gova"
              className="gova-btn gova-btn--ghost"
              target="_blank"
              rel="noreferrer"
            >
              <GithubGlyph /> GitHub
            </Link>
          </div>

          <div
            className="gova-rise"
            style={{
              animationDelay: '300ms',
              display: 'flex',
              flexWrap: 'wrap',
              alignItems: 'center',
              gap: '0.85rem',
              fontFamily: 'var(--font-mono)',
              fontSize: '0.72rem',
              letterSpacing: '0.06em',
              textTransform: 'uppercase',
              color: 'var(--gova-cream-faint)',
              marginTop: '0.1rem',
            }}
          >
            <span>Go 1.26+</span>
            <span style={metaDot} />
            <span>macOS</span>
            <span style={metaDot} />
            <span>Windows</span>
            <span style={metaDot} />
            <span>Linux</span>
          </div>
        </div>

        <div
          className="gova-rise gova-hero-code-col"
          style={{ animationDelay: '320ms', minWidth: 0 }}
        >
          <HeroCode />
        </div>
      </div>
    </section>
  );
}

function InstallCommand() {
  return (
    <div
      className="gova-rise"
      style={{
        animationDelay: '180ms',
        display: 'flex',
        alignItems: 'center',
        gap: '0.65rem',
        padding: '0.6rem 0.75rem 0.6rem 0.9rem',
        background: 'var(--gova-ink-2)',
        border: '1px solid var(--gova-line)',
        borderRadius: 10,
        fontFamily: 'var(--font-mono)',
        fontSize: '0.85rem',
        color: 'var(--gova-cream)',
        maxWidth: 480,
        width: '100%',
      }}
    >
      <span style={{ color: 'var(--gova-cream-faint)', userSelect: 'none' }}>$</span>
      <span
        style={{
          flex: 1,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
        }}
      >
        go get github.com/nv404/gova@latest
      </span>
      <span
        style={{
          fontSize: '0.66rem',
          padding: '0.18rem 0.5rem',
          borderRadius: 5,
          background: 'var(--gova-surface)',
          border: '1px solid var(--gova-line)',
          color: 'var(--gova-cream-dim)',
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
        }}
      >
        shell
      </span>
    </div>
  );
}

function HeroCode() {
  return (
    <CodeWindow
      title="main.go"
      lines={[
        <>
          <Kw>package</Kw> <Pk>main</Pk>
        </>,
        <>&nbsp;</>,
        <>
          <Kw>import</Kw> . <Str>&quot;github.com/nv404/gova&quot;</Str>
        </>,
        <>&nbsp;</>,
        <>
          <Kw>var</Kw> <Ty>Counter</Ty> = <Fn>Define</Fn>(<Kw>func</Kw>(s *<Ty>Scope</Ty>) <Ty>View</Ty> {'{'}
        </>,
        <>
          {'  '}count := <Fn>State</Fn>(s, <Nm>0</Nm>)
        </>,
        <>&nbsp;</>,
        <>
          {'  '}<Kw>return</Kw> <Fn>VStack</Fn>(
        </>,
        <>
          {'    '}<Fn>Text</Fn>(count.<Fn>Format</Fn>(<Str>&quot;Count: %d&quot;</Str>)).<Fn>Font</Fn>(<Ty>Title</Ty>),
        </>,
        <>
          {'    '}<Fn>Button</Fn>(<Str>&quot;+&quot;</Str>, <Kw>func</Kw>() {'{'} count.<Fn>Update</Fn>(<Kw>func</Kw>(n <Ty>int</Ty>) <Ty>int</Ty> {'{'} <Kw>return</Kw> n + <Nm>1</Nm> {'}'}) {'}'}),
        </>,
        <>
          {'    '}<Fn>Button</Fn>(<Str>&quot;-&quot;</Str>, <Kw>func</Kw>() {'{'} count.<Fn>Update</Fn>(<Kw>func</Kw>(n <Ty>int</Ty>) <Ty>int</Ty> {'{'} <Kw>return</Kw> n - <Nm>1</Nm> {'}'}) {'}'}),
        </>,
        <>
          {'  '})
        </>,
        <>{'})'}</>,
        <>&nbsp;</>,
        <>
          <Kw>func</Kw> <Fn>main</Fn>() {'{'} <Fn>Run</Fn>(<Str>&quot;Counter&quot;</Str>, Counter) {'}'}
        </>,
      ]}
    />
  );
}

/* ---------- Metrics ---------- */

function Metrics() {
  const stats = [
    { label: 'Binary size', value: '32 MB', hint: 'counter, default go build' },
    { label: 'Stripped', value: '23 MB', hint: 'go build -ldflags "-s -w"' },
    { label: 'Memory idle', value: '~80 MB', hint: 'RSS, counter running' },
    { label: 'Go version', value: '1.26+', hint: 'plus C toolchain for cgo' },
    { label: 'License', value: 'MIT', hint: 'no runtime fees' },
  ];
  return (
    <section
      style={{
        borderBottom: '1px solid var(--gova-line)',
        background: 'var(--gova-ink-2)',
      }}
    >
      <div
        style={{
          maxWidth: 1200,
          margin: '0 auto',
          padding: 'clamp(2rem, 4vw, 2.75rem) clamp(1.25rem, 4vw, 2.5rem)',
        }}
      >
        <div
          className="gova-metrics-grid"
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(5, minmax(0, 1fr))',
            gap: 'clamp(1rem, 2vw, 1.75rem)',
          }}
        >
          {stats.map((s) => (
            <div
              key={s.label}
              style={{ display: 'flex', flexDirection: 'column', gap: 6 }}
            >
              <span
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.68rem',
                  letterSpacing: '0.12em',
                  textTransform: 'uppercase',
                  color: 'var(--gova-cream-faint)',
                }}
              >
                {s.label}
              </span>
              <span
                style={{
                  fontFamily: 'var(--font-sans)',
                  fontSize: 'clamp(1.35rem, 2.3vw, 1.75rem)',
                  fontWeight: 500,
                  letterSpacing: '-0.025em',
                  color: 'var(--gova-cream)',
                  lineHeight: 1.05,
                }}
              >
                {s.value}
              </span>
              <span
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.7rem',
                  color: 'var(--gova-cream-faint)',
                  lineHeight: 1.45,
                }}
              >
                {s.hint}
              </span>
            </div>
          ))}
        </div>
        <p
          style={{
            marginTop: '1.4rem',
            marginBottom: 0,
            fontFamily: 'var(--font-mono)',
            fontSize: '0.72rem',
            color: 'var(--gova-cream-faint)',
            letterSpacing: '0.02em',
          }}
        >
          Measured on macOS arm64 with Go 1.26.2, counter example, static build.
          Your numbers will vary by platform and feature set.
        </p>
      </div>
    </section>
  );
}

/* ---------- Features ---------- */

function Features() {
  const features = [
    {
      n: '01',
      title: 'Components as structs',
      body:
        'Views are plain Go structs. Props are fields, defaults are zero values, and composition is just function calls. The compiler checks your UI.',
    },
    {
      n: '02',
      title: 'Explicit reactive scope',
      body:
        'State, signals, and effects live on a Scope you can see. No hidden scheduler, no hook-order rules, no re-render surprises.',
    },
    {
      n: '03',
      title: 'Real native dialogs',
      body:
        'NSAlert, NSOpenPanel, and NSDockTile on macOS through cgo. Fyne fallbacks on Windows and Linux. Same API everywhere.',
    },
    {
      n: '04',
      title: 'One static binary',
      body:
        'go build produces a single executable. No JavaScript runtime, no embedded browser, no extra assets to bundle.',
    },
    {
      n: '05',
      title: 'Hot reload in dev',
      body:
        'gova dev watches Go files and rebuilds on save. Ignores .git, node_modules, vendor, and _test.go files by default.',
    },
    {
      n: '06',
      title: 'Cross-platform by default',
      body:
        'One codebase compiles to macOS, Windows, and Linux. Platform-specific code degrades to safe fallbacks so your app keeps running.',
    },
  ];
  return (
    <Section title="What the framework gives you." width={1200}>
      <div
        className="gova-features-grid"
        style={{
          marginTop: '2.25rem',
          display: 'grid',
          gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
          gap: '1px',
          background: 'var(--gova-line)',
          border: '1px solid var(--gova-line)',
          borderRadius: 14,
          overflow: 'hidden',
        }}
      >
        {features.map((f) => (
          <article
            key={f.n}
            className="gova-feature-cell"
            style={{
              padding: '1.65rem 1.5rem 1.55rem',
              background: 'var(--gova-ink)',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.75rem',
              minHeight: 210,
              transition: 'background 200ms ease',
            }}
          >
            <div
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '0.68rem',
                letterSpacing: '0.14em',
                color: 'var(--gova-cream-faint)',
              }}
            >
              {f.n} / 06
            </div>
            <h3
              style={{
                margin: 0,
                fontFamily: 'var(--font-sans)',
                fontWeight: 500,
                fontSize: '1.08rem',
                letterSpacing: '-0.012em',
                color: 'var(--gova-cream)',
              }}
            >
              {f.title}
            </h3>
            <p
              style={{
                margin: 0,
                fontSize: '0.92rem',
                lineHeight: 1.55,
                color: 'var(--gova-cream-dim)',
              }}
            >
              {f.body}
            </p>
          </article>
        ))}
      </div>
    </Section>
  );
}

/* ---------- Platforms ---------- */

type PlatformState = 'full' | 'partial' | 'planned';

function Platforms() {
  const rows: Array<[string, PlatformState, PlatformState, PlatformState, string]> = [
    ['Core UI: layouts, widgets, navigation', 'full', 'full', 'full', ''],
    ['Hot reload via gova dev', 'full', 'full', 'full', ''],
    ['App icon at runtime', 'full', 'full', 'full', ''],
    [
      'Native dialogs',
      'full',
      'partial',
      'partial',
      'NSAlert and NSOpenPanel on macOS. Fyne fallback elsewhere.',
    ],
    [
      'Dock and taskbar (badge, progress, menu)',
      'full',
      'planned',
      'planned',
      'NSDockTile on macOS. Planned on Windows and Linux.',
    ],
  ];
  return (
    <Section title="Platform support." width={1200}>
      <p style={sectionSub}>
        Gova compiles to macOS, Windows, and Linux from the same Go file.
        Platform integrations degrade to safe fallbacks so portable code stays
        portable.
      </p>

      <div
        style={{
          marginTop: '2.25rem',
          border: '1px solid var(--gova-line)',
          borderRadius: 14,
          overflow: 'hidden',
          background:
            'linear-gradient(180deg, var(--gova-surface) 0%, var(--gova-ink-2) 100%)',
        }}
      >
        <div
          className="gova-platform-header"
          style={{
            display: 'grid',
            gridTemplateColumns: 'minmax(0, 2fr) repeat(3, minmax(0, 1fr))',
            padding: '0.9rem 1.15rem',
            borderBottom: '1px solid var(--gova-line)',
            background: 'rgba(247,244,229,0.035)',
            fontFamily: 'var(--font-mono)',
            fontSize: '0.7rem',
            letterSpacing: '0.14em',
            textTransform: 'uppercase',
            color: 'var(--gova-cream-faint)',
          }}
        >
          <span>Feature</span>
          <span>macOS</span>
          <span>Windows</span>
          <span>Linux</span>
        </div>
        {rows.map((r, i) => (
          <div
            key={i}
            className="gova-platform-row"
            style={{
              display: 'grid',
              gridTemplateColumns: 'minmax(0, 2fr) repeat(3, minmax(0, 1fr))',
              padding: '1rem 1.15rem',
              borderTop: i === 0 ? 'none' : '1px solid var(--gova-line)',
              alignItems: 'center',
              gap: '0.75rem',
            }}
          >
            <div>
              <div style={{ color: 'var(--gova-cream)', fontSize: '0.94rem' }}>
                {r[0]}
              </div>
              {r[4] ? (
                <div
                  style={{
                    marginTop: 4,
                    fontFamily: 'var(--font-mono)',
                    fontSize: '0.7rem',
                    color: 'var(--gova-cream-faint)',
                  }}
                >
                  {r[4]}
                </div>
              ) : null}
            </div>
            <div className="gova-platform-cell" data-label="macOS">
              <StatusPill s={r[1]} />
            </div>
            <div className="gova-platform-cell" data-label="Windows">
              <StatusPill s={r[2]} />
            </div>
            <div className="gova-platform-cell" data-label="Linux">
              <StatusPill s={r[3]} />
            </div>
          </div>
        ))}
      </div>
    </Section>
  );
}

function StatusPill({ s }: { s: PlatformState }) {
  const conf = {
    full: {
      label: 'Supported',
      dot: 'var(--gova-moss)',
      bg: 'rgba(168,181,142,0.12)',
      fg: 'var(--gova-moss)',
    },
    partial: {
      label: 'Partial',
      dot: 'var(--gova-cream-dim)',
      bg: 'rgba(247,244,229,0.08)',
      fg: 'var(--gova-cream)',
    },
    planned: {
      label: 'Planned',
      dot: 'var(--gova-cream-faint)',
      bg: 'rgba(115,112,95,0.18)',
      fg: 'var(--gova-cream-dim)',
    },
  }[s];
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 8,
        fontFamily: 'var(--font-mono)',
        fontSize: '0.74rem',
        padding: '0.3rem 0.7rem',
        borderRadius: 999,
        background: conf.bg,
        color: conf.fg,
        alignSelf: 'flex-start',
        width: 'fit-content',
      }}
    >
      <span
        style={{ width: 6, height: 6, borderRadius: 999, background: conf.dot }}
      />
      {conf.label}
    </span>
  );
}

/* ---------- Toolchain ---------- */

function Toolchain() {
  const cmds = [
    {
      cmd: 'gova dev',
      purpose: 'Hot reload',
      body:
        'Watch .go files in the working directory, rebuild on save, and relaunch the window. Ignores .git, node_modules, vendor, and _test.go.',
    },
    {
      cmd: 'gova build',
      purpose: 'Compile',
      body:
        'Produce a static binary for the current platform. Pair with -ldflags "-s -w" to strip debug info for a smaller artifact.',
    },
    {
      cmd: 'gova run',
      purpose: 'Execute',
      body:
        'Build and launch the app once, without file watching. Useful for CI smoke tests or one-off demos.',
    },
  ];
  return (
    <Section title="Command-line tools." width={1200}>
      <p style={sectionSub}>
        The gova CLI ships alongside the framework. Install it with{' '}
        <span style={monoInline}>
          go install github.com/nv404/gova/cmd/gova@latest
        </span>
        .
      </p>
      <div
        className="gova-value-grid"
        style={{
          marginTop: '2.25rem',
          display: 'grid',
          gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
          gap: '1rem',
        }}
      >
        {cmds.map((c) => (
          <div
            key={c.cmd}
            className="gova-card gova-reveal"
            style={{
              padding: '1.4rem 1.35rem',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.85rem',
            }}
          >
            <div
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '0.68rem',
                letterSpacing: '0.14em',
                textTransform: 'uppercase',
                color: 'var(--gova-cream-faint)',
              }}
            >
              {c.purpose}
            </div>
            <code
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '0.98rem',
                color: 'var(--gova-cream)',
                background: 'var(--gova-ink-2)',
                border: '1px solid var(--gova-line)',
                borderRadius: 6,
                padding: '0.38rem 0.65rem',
                alignSelf: 'flex-start',
              }}
            >
              {c.cmd}
            </code>
            <p
              style={{
                margin: 0,
                fontSize: '0.9rem',
                lineHeight: 1.55,
                color: 'var(--gova-cream-dim)',
              }}
            >
              {c.body}
            </p>
          </div>
        ))}
      </div>
    </Section>
  );
}

/* ---------- Quickstart ---------- */

function Quickstart() {
  const steps = [
    {
      n: '01',
      title: 'Install the prerequisites',
      body:
        'Go 1.26 or later, plus a C toolchain for cgo. On macOS: Xcode Command Line Tools. On Linux: build-essential and libgl1-mesa-dev. On Windows: mingw-w64.',
      code: null,
    },
    {
      n: '02',
      title: 'Pull the module',
      body: 'Add gova to a new or existing Go module.',
      code: 'go mod init myapp\ngo get github.com/nv404/gova',
    },
    {
      n: '03',
      title: 'Run the counter example',
      body:
        'Clone the repo and run the example directly. No scaffold, no boilerplate.',
      code:
        'git clone https://github.com/nv404/gova\ncd gova\ngo run ./examples/counter',
    },
  ];
  return (
    <Section title="Quickstart." width={1200}>
      <div
        style={{
          marginTop: '2.25rem',
          display: 'flex',
          flexDirection: 'column',
          gap: '1rem',
        }}
      >
        {steps.map((s) => (
          <div
            key={s.n}
            className="gova-quickstart-row gova-reveal"
            style={{
              display: 'grid',
              gridTemplateColumns: '68px minmax(0, 1fr) minmax(0, 1.25fr)',
              gap: '1.5rem',
              padding: '1.35rem 1.5rem',
              border: '1px solid var(--gova-line)',
              borderRadius: 12,
              background:
                'linear-gradient(180deg, var(--gova-surface) 0%, var(--gova-ink-2) 100%)',
              alignItems: 'center',
            }}
          >
            <div
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '1.3rem',
                color: 'var(--gova-cream)',
                letterSpacing: '0.02em',
              }}
            >
              {s.n}
            </div>
            <div>
              <h4
                style={{
                  margin: 0,
                  fontFamily: 'var(--font-sans)',
                  fontWeight: 500,
                  fontSize: '1.05rem',
                  letterSpacing: '-0.01em',
                  color: 'var(--gova-cream)',
                }}
              >
                {s.title}
              </h4>
              <p
                style={{
                  margin: '0.45rem 0 0 0',
                  fontSize: '0.9rem',
                  lineHeight: 1.55,
                  color: 'var(--gova-cream-dim)',
                }}
              >
                {s.body}
              </p>
            </div>
            {s.code ? (
              <pre
                style={{
                  margin: 0,
                  padding: '0.85rem 1rem',
                  background: 'var(--gova-ink)',
                  border: '1px solid var(--gova-line)',
                  borderRadius: 8,
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.8rem',
                  color: 'var(--gova-cream)',
                  overflowX: 'auto',
                  lineHeight: 1.65,
                }}
              >
                <code>{s.code}</code>
              </pre>
            ) : (
              <div
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.78rem',
                  color: 'var(--gova-cream-faint)',
                  lineHeight: 1.55,
                }}
              >
                See the{' '}
                <Link
                  href="/docs/getting-started/installation"
                  style={{
                    color: 'var(--gova-cream)',
                    textDecoration: 'underline',
                    textUnderlineOffset: 3,
                  }}
                >
                  installation guide
                </Link>{' '}
                for platform specifics.
              </div>
            )}
          </div>
        ))}
      </div>
    </Section>
  );
}

/* ---------- Final CTA ---------- */

function CTA() {
  return (
    <section
      style={{
        position: 'relative',
        padding: 'clamp(4rem, 9vw, 7rem) 1.5rem',
        textAlign: 'center',
        borderTop: '1px solid var(--gova-line)',
        marginTop: '3rem',
        overflow: 'hidden',
      }}
    >
      <div
        aria-hidden
        style={{
          position: 'absolute',
          inset: 0,
          background:
            'radial-gradient(ellipse at 50% 120%, rgba(247,244,229,0.1), transparent 60%)',
          pointerEvents: 'none',
        }}
      />
      <div style={{ position: 'relative', maxWidth: 760, margin: '0 auto' }}>
        <h2
          style={{
            fontFamily: 'var(--font-sans)',
            fontWeight: 500,
            fontSize: 'clamp(2rem, 4.6vw, 3.2rem)',
            lineHeight: 1.05,
            letterSpacing: '-0.035em',
            margin: 0,
            color: 'var(--gova-cream)',
          }}
        >
          Try it. Break it. File an issue.
        </h2>
        <p
          style={{
            marginTop: '1.25rem',
            fontSize: '1rem',
            lineHeight: 1.6,
            color: 'var(--gova-cream-dim)',
            maxWidth: 560,
            margin: '1.25rem auto 0',
          }}
        >
          Gova is pre-1.0 and the API is still moving. If you build a Go desktop
          app with it, we want to hear what worked and what did not.
        </p>
        <div
          style={{
            marginTop: '2rem',
            display: 'flex',
            gap: '0.65rem',
            justifyContent: 'center',
            flexWrap: 'wrap',
          }}
        >
          <Link
            href="/docs/getting-started/installation"
            className="gova-btn gova-btn--primary"
          >
            Get started <Arrow />
          </Link>
          <Link
            href="https://github.com/nv404/gova"
            className="gova-btn gova-btn--ghost"
            target="_blank"
            rel="noreferrer"
          >
            <GithubGlyph /> Star on GitHub
          </Link>
        </div>
      </div>
    </section>
  );
}

/* ---------- Footer ---------- */

function Footer() {
  return (
    <footer
      style={{
        borderTop: '1px solid var(--gova-line)',
        padding: '2.25rem 1.5rem',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        gap: '1rem',
        flexWrap: 'wrap',
        maxWidth: 1200,
        margin: '0 auto',
      }}
    >
      <div
        style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '0.76rem',
          color: 'var(--gova-cream-faint)',
          letterSpacing: '0.02em',
        }}
      >
        © {new Date().getFullYear()} Gova · MIT · built with Go and cgo
      </div>
      <div style={{ display: 'flex', gap: '1.25rem' }}>
        <FootLink href="/docs">Docs</FootLink>
        <FootLink href="https://github.com/nv404/gova/tree/main/examples">
          Examples
        </FootLink>
        <FootLink href="https://github.com/nv404/gova">GitHub</FootLink>
        <FootLink href="https://github.com/nv404/gova/issues">Issues</FootLink>
      </div>
    </footer>
  );
}

function FootLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <Link
      href={href}
      style={{
        fontFamily: 'var(--font-mono)',
        fontSize: '0.76rem',
        color: 'var(--gova-cream-dim)',
        textDecoration: 'none',
        letterSpacing: '0.02em',
      }}
    >
      {children}
    </Link>
  );
}

/* ---------- Shared primitives ---------- */

function Section({
  title,
  width = 1200,
  children,
}: {
  title: string;
  width?: number;
  children: ReactNode;
}) {
  return (
    <section
      style={{
        position: 'relative',
        maxWidth: width,
        margin: '0 auto',
        padding: 'clamp(3.25rem, 7vw, 5.5rem) clamp(1.25rem, 4vw, 2.5rem)',
      }}
    >
      <h2
        style={{
          margin: 0,
          fontFamily: 'var(--font-sans)',
          fontWeight: 500,
          fontSize: 'clamp(1.7rem, 3.6vw, 2.4rem)',
          lineHeight: 1.1,
          letterSpacing: '-0.028em',
          color: 'var(--gova-cream)',
          maxWidth: 780,
        }}
      >
        {title}
      </h2>
      {children}
    </section>
  );
}

function CodeWindow({
  title,
  lines,
}: {
  title: string;
  lines: ReactNode[];
}) {
  return (
    <div
      className="gova-code"
      style={{
        boxShadow: '0 30px 70px -30px rgba(0,0,0,0.7)',
      }}
    >
      {title ? (
        <div
          style={{
            padding: '0.55rem 0.85rem',
            borderBottom: '1px solid var(--gova-line)',
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            background: 'rgba(15,14,13,0.5)',
          }}
        >
          <span style={traffic('#FF5F56')} />
          <span style={traffic('#FFBD2E')} />
          <span style={traffic('#27C93F')} />
          <span
            style={{
              marginLeft: 10,
              fontFamily: 'var(--font-mono)',
              fontSize: '0.72rem',
              color: 'var(--gova-cream-faint)',
            }}
          >
            {title}
          </span>
        </div>
      ) : null}
      <pre
        style={{
          margin: 0,
          padding: '1.1rem 1.2rem',
          overflowX: 'auto',
          fontFamily: 'var(--font-mono)',
          counterReset: 'line',
        }}
      >
        <code>
          {lines.map((l, i) => (
            <div
              key={i}
              style={{
                display: 'grid',
                gridTemplateColumns: '2.2em 1fr',
                gap: '0.25rem',
              }}
            >
              <span
                style={{
                  color: 'var(--gova-cream-faint)',
                  opacity: 0.5,
                  userSelect: 'none',
                  fontSize: '0.72rem',
                  textAlign: 'right',
                  paddingRight: '0.4rem',
                  lineHeight: 1.7,
                }}
              >
                {i + 1}
              </span>
              <span>{l}</span>
            </div>
          ))}
        </code>
      </pre>
    </div>
  );
}

/* --- Syntax token helpers --- */
function Kw({ children }: { children: ReactNode }) {
  return <span className="tok-kw">{children}</span>;
}
function Str({ children }: { children: ReactNode }) {
  return <span className="tok-str">{children}</span>;
}
function Fn({ children }: { children: ReactNode }) {
  return <span className="tok-fn">{children}</span>;
}
function Ty({ children }: { children: ReactNode }) {
  return <span className="tok-ty">{children}</span>;
}
function Pk({ children }: { children: ReactNode }) {
  return <span className="tok-pk">{children}</span>;
}
function Nm({ children }: { children: ReactNode }) {
  return <span className="tok-nm">{children}</span>;
}

/* --- Glyphs --- */
function Arrow() {
  return (
    <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden>
      <path
        d="M2 7h10M8 3l4 4-4 4"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
function GithubGlyph() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
      <path d="M12 .5C5.65.5.5 5.65.5 12a11.5 11.5 0 0 0 7.86 10.92c.58.1.79-.25.79-.56v-2.1c-3.2.7-3.88-1.36-3.88-1.36-.53-1.33-1.3-1.68-1.3-1.68-1.05-.72.08-.7.08-.7 1.17.08 1.78 1.2 1.78 1.2 1.04 1.78 2.73 1.27 3.4.97.1-.76.4-1.27.74-1.56-2.56-.3-5.26-1.28-5.26-5.72 0-1.26.45-2.3 1.2-3.1-.13-.3-.52-1.5.1-3.13 0 0 .98-.32 3.2 1.18a11.1 11.1 0 0 1 5.82 0c2.22-1.5 3.2-1.18 3.2-1.18.62 1.63.23 2.83.1 3.12.75.82 1.2 1.85 1.2 3.12 0 4.45-2.7 5.41-5.28 5.7.42.37.78 1.08.78 2.18v3.22c0 .31.2.67.8.56A11.5 11.5 0 0 0 23.5 12C23.5 5.65 18.35.5 12 .5Z" />
    </svg>
  );
}

/* --- CSS-in-JS helpers --- */
function traffic(color: string): CSSProperties {
  return {
    width: 10,
    height: 10,
    borderRadius: 999,
    background: color,
    display: 'inline-block',
  };
}

const metaDot: CSSProperties = {
  width: 3,
  height: 3,
  borderRadius: 999,
  background: 'var(--gova-cream-faint)',
  display: 'inline-block',
};

const monoInline: CSSProperties = {
  fontFamily: 'var(--font-mono)',
  fontSize: '0.9em',
  color: 'var(--gova-cream)',
  padding: '0.08rem 0.35rem',
  background: 'var(--gova-ink-2)',
  border: '1px solid var(--gova-line)',
  borderRadius: 4,
};

const sectionSub: CSSProperties = {
  maxWidth: 680,
  marginTop: '1.1rem',
  marginBottom: 0,
  fontSize: '1rem',
  lineHeight: 1.65,
  color: 'var(--gova-cream-dim)',
};
