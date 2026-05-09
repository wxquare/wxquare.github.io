// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

(() => {
    const darkThemes = ['ayu', 'navy', 'coal'];
    const lightThemes = ['light', 'rust'];

    const convertMermaidCodeBlocks = () => {
        for (const code of document.querySelectorAll('pre > code.language-mermaid')) {
            const pre = code.parentElement;
            if (!pre || pre.classList.contains('mermaid')) {
                continue;
            }
            pre.className = 'mermaid';
            pre.textContent = code.textContent;
        }
    };

    const classList = document.getElementsByTagName('html')[0].classList;

    let lastThemeWasLight = true;
    for (const cssClass of classList) {
        if (darkThemes.includes(cssClass)) {
            lastThemeWasLight = false;
            break;
        }
    }

    const theme = lastThemeWasLight ? 'default' : 'dark';

    const renderMermaid = () => {
        convertMermaidCodeBlocks();

        const nodes = document.querySelectorAll('.mermaid');
        if (nodes.length === 0) {
            return;
        }

        mermaid.initialize({ startOnLoad: false, theme });

        if (typeof mermaid.run === 'function') {
            mermaid.run({ nodes });
            return;
        }

        mermaid.init(undefined, nodes);
    };

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', renderMermaid);
    } else {
        renderMermaid();
    }

    // Simplest way to make mermaid re-render the diagrams in the new theme is via refreshing the page

    for (const darkTheme of darkThemes) {
        document.getElementById(darkTheme)?.addEventListener('click', () => {
            if (lastThemeWasLight) {
                window.location.reload();
            }
        });
    }

    for (const lightTheme of lightThemes) {
        document.getElementById(lightTheme)?.addEventListener('click', () => {
            if (!lastThemeWasLight) {
                window.location.reload();
            }
        });
    }
})();
