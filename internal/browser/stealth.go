package browser

// StealthInitScript é o script injetado antes de qualquer página carregar (page.AddInitScript)
// para mascarar características de automação do Chromium e passar por WAFs (F5, Cloudflare, etc).
const StealthInitScript = `
(() => {
    // 1. Ocultar navigator.webdriver
    Object.defineProperty(navigator, 'webdriver', {
        get: () => undefined,
        configurable: true
    });

    // 2. Simular window.chrome
    if (!window.chrome) {
        window.chrome = {
            app: {
                isInstalled: false,
                InstallState: { DISABLED: 'DISABLED', INSTALLED: 'INSTALLED', NOT_INSTALLED: 'NOT_INSTALLED' },
                RunningState: { CANNOT_RUN: 'CANNOT_RUN', READY_TO_RUN: 'READY_TO_RUN', RUNNING: 'RUNNING' }
            },
            runtime: {
                OnInstalledReason: { CHROME_UPDATE: 'chrome_update', INSTALL: 'install', SHARED_MODULE_UPDATE: 'shared_module_update', UPDATE: 'update' },
                OnRestartRequiredReason: { APP_UPDATE: 'app_update', OS_UPDATE: 'os_update', PERIODIC: 'periodic' },
                PlatformArch: { ARM: 'arm', ARM64: 'arm64', MIPS: 'mips', MIPS64: 'mips64', X86_32: 'x86-32', X86_64: 'x86-64' },
                PlatformNaclArch: { ARM: 'arm', MIPS: 'mips', MIPS64: 'mips64', X86_32: 'x86-32', X86_64: 'x86-64' },
                PlatformOs: { ANDROID: 'android', CROS: 'cros', LINUX: 'linux', MAC: 'mac', OPENBSD: 'openbsd', WIN: 'win' },
                RequestUpdateCheckStatus: { NO_UPDATE: 'no_update', THROTTLED: 'throttled', UPDATE_AVAILABLE: 'update_available' }
            }
        };
    }

    // 3. Simular plugins válidos no navigator
    Object.defineProperty(navigator, 'plugins', {
        get: () => {
            const plugins = [
                { name: 'PDF Viewer', filename: 'internal-pdf-viewer', description: 'Portable Document Format' },
                { name: 'Chrome PDF Viewer', filename: 'internal-pdf-viewer', description: 'Portable Document Format' },
                { name: 'Chromium PDF Viewer', filename: 'internal-pdf-viewer', description: 'Portable Document Format' }
            ];
            plugins.item = (i) => plugins[i] || null;
            plugins.namedItem = (name) => plugins.find(p => p.name === name) || null;
            plugins.refresh = () => {};
            return plugins;
        },
        configurable: true
    });

    // 4. Idiomas padrão Brasil
    Object.defineProperty(navigator, 'languages', {
        get: () => ['pt-BR', 'pt', 'en-US', 'en'],
        configurable: true
    });

    // 5. Mock da API de permissões para Notifications
    const originalQuery = window.navigator.permissions.query;
    window.navigator.permissions.query = (parameters) => (
        parameters.name === 'notifications' ?
            Promise.resolve({ state: Notification.permission }) :
            originalQuery(parameters)
    );
})();
`

// DefaultUserAgent padrão atualizado de Chrome em Linux Desktop
const DefaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
