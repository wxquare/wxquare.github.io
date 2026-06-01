// Populate the sidebar
//
// This is a script, and not included directly in the page, to control the total size of the book.
// The TOC contains an entry for each page, so if each page includes a copy of the TOC,
// the total size of the page becomes O(n**2).
class MDBookSidebarScrollbox extends HTMLElement {
    constructor() {
        super();
    }
    connectedCallback() {
        this.innerHTML = '<ol class="chapter"><li class="chapter-item "><span class="chapter-link-wrapper"><a href="index.html">前言与使用说明</a></span></li><li class="chapter-item "><li class="spacer"></li></li><li class="chapter-item "><li class="part-title">第一部分：系统设计方法论</li></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/01-system-design-guide.html"><strong aria-hidden="true">1.</strong> 第 1 章 系统设计完全指南</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/02-technical-design-methodology.html"><strong aria-hidden="true">2.</strong> 第 2 章 技术方案设计方法论</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/03-architecture-combination.html"><strong aria-hidden="true">3.</strong> 第 3 章 架构师的组合拳</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/04-business-boundary-strategic-design.html"><strong aria-hidden="true">4.</strong> 第 4 章 业务边界与战略设计</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/05-internal-architecture-design.html"><strong aria-hidden="true">5.</strong> 第 5 章 系统内部结构设计</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/06-integration-consistency-design.html"><strong aria-hidden="true">6.</strong> 第 6 章 系统集成与一致性设计</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/07-architecture-quality-assurance.html"><strong aria-hidden="true">7.</strong> 第 7 章 架构质量保障</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/08-coding-principles-design-patterns.html"><strong aria-hidden="true">8.</strong> 第 8 章 编码原则与设计模式</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/09-system-reliability-engineering.html"><strong aria-hidden="true">9.</strong> 第 9 章 系统可靠性工程</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/10-reconciliation-compensation-dlq.html"><strong aria-hidden="true">10.</strong> 第 10 章 对账、补偿、DLQ 与故障恢复</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/11-asset-loss-prevention.html"><strong aria-hidden="true">11.</strong> 第 11 章 资损防控</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part01/12-capacity-planning-resilience.html"><strong aria-hidden="true">12.</strong> 第 12 章 容量规划、压测、限流、熔断与降级</a></span></li><li class="chapter-item "><li class="spacer"></li></li><li class="chapter-item "><li class="part-title">第二部分：电商系统设计实战</li></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/01-ecommerce-overview.html"><strong aria-hidden="true">13.</strong> 第 20 章 电商系统全景图</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/02-product-center.html"><strong aria-hidden="true">14.</strong> 第 21 章 商品中心系统</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/03-inventory-system.html"><strong aria-hidden="true">15.</strong> 第 22 章 库存系统</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/04-marketing-system.html"><strong aria-hidden="true">16.</strong> 第 23 章 营销系统</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/05-product-supply-ops.html"><strong aria-hidden="true">17.</strong> 第 24 章 商品供给管理</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/06-pricing-system.html"><strong aria-hidden="true">18.</strong> 第 25 章 计价系统设计与实现</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/07-search-discovery.html"><strong aria-hidden="true">19.</strong> 第 26 章 搜索与导购</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/08-cart-checkout.html"><strong aria-hidden="true">20.</strong> 第 27 章 购物车与结算</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/09-order-system.html"><strong aria-hidden="true">21.</strong> 第 28 章 订单系统</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/10-payment-system.html"><strong aria-hidden="true">22.</strong> 第 29 章 支付系统</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/11-supplier-sync.html"><strong aria-hidden="true">23.</strong> 第 30 章 供应商数据同步链路</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/12-product-supply-governance.html"><strong aria-hidden="true">24.</strong> 第 31 章 商品供给与运营治理平台</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part02/13-b2b2c-platform-architecture.html"><strong aria-hidden="true">25.</strong> 第 32 章 B2B2C 平台完整架构</a></span></li><li class="chapter-item "><li class="spacer"></li></li><li class="chapter-item "><li class="part-title">第三部分：系统设计面试</li></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part03/01-system-design-interview-overview.html"><strong aria-hidden="true">26.</strong> 第 33 章 系统设计面试综合</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part03/02-middleware-reliability-interview.html"><strong aria-hidden="true">27.</strong> 第 34 章 中间件与可靠性高频追问</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part03/03-ecommerce-architecture-interview.html"><strong aria-hidden="true">28.</strong> 第 35 章 电商架构面试题精选</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part03/04-product-inventory-marketing-pricing-interview.html"><strong aria-hidden="true">29.</strong> 第 36 章 商品、库存、营销与计价专题</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part03/05-search-cart-order-payment-interview.html"><strong aria-hidden="true">30.</strong> 第 37 章 搜索、购物车、订单与支付专题</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part03/06-whiteboard-capacity-estimation.html"><strong aria-hidden="true">31.</strong> 第 38 章 白板答辩与容量估算表达</a></span></li><li class="chapter-item "><li class="spacer"></li></li><li class="chapter-item "><li class="part-title">第四部分：基础设施与计算机基础</li></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/01-mysql-storage-database.html"><strong aria-hidden="true">32.</strong> 第 9 章 MySQL：存储与数据库</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/02-redis-cache-practice.html"><strong aria-hidden="true">33.</strong> 第 10 章 Redis：缓存原理与实践</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/03-kafka-message-queue-async.html"><strong aria-hidden="true">34.</strong> 第 11 章 Kafka：消息队列与异步</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/04-elasticsearch-search-index.html"><strong aria-hidden="true">35.</strong> 第 12 章 Elasticsearch：搜索与索引</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/05-kubernetes-docker.html"><strong aria-hidden="true">36.</strong> 第 13 章 Kubernetes 与 Docker</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/06-global-id-and-basic-services.html"><strong aria-hidden="true">37.</strong> 第 14 章 全局 ID 体系与基础服务设计</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/07-tech-stack-selection.html"><strong aria-hidden="true">38.</strong> 第 15 章 技术栈选型指南</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/08-operating-system.html"><strong aria-hidden="true">39.</strong> 第 39 章 操作系统基础</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/09-computer-networking.html"><strong aria-hidden="true">40.</strong> 第 40 章 计算机网络实践</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/10-bash-shell-practice.html"><strong aria-hidden="true">41.</strong> 第 41 章 Bash 与 Shell 实用</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/11-python-practice.html"><strong aria-hidden="true">42.</strong> 第 42 章 Python 实践</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/12-cpp-practice.html"><strong aria-hidden="true">43.</strong> 第 43 章 C++ 实践</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/13-go-practice.html"><strong aria-hidden="true">44.</strong> 第 44 章 Go 语言实践</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="part04/14-data-structures-and-algorithms.html"><strong aria-hidden="true">45.</strong> 第 45 章 数据结构与算法题型速查</a></span></li><li class="chapter-item "><li class="spacer"></li></li><li class="chapter-item "><li class="part-title">附录</li></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="appendix/glossary.html"><strong aria-hidden="true">46.</strong> 附录 A 术语表</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="appendix/references.html"><strong aria-hidden="true">47.</strong> 附录 B 参考文献与外链</a></span></li><li class="chapter-item "><span class="chapter-link-wrapper"><a href="appendix/tooling.html"><strong aria-hidden="true">48.</strong> 附录 C 工具与构建说明</a></span></li></ol>';
        // Set the current, active page, and reveal it if it's hidden
        let current_page = document.location.href.toString().split('#')[0].split('?')[0];
        if (current_page.endsWith('/')) {
            current_page += 'index.html';
        }
        const links = Array.prototype.slice.call(this.querySelectorAll('a'));
        const l = links.length;
        for (let i = 0; i < l; ++i) {
            const link = links[i];
            const href = link.getAttribute('href');
            if (href && !href.startsWith('#') && !/^(?:[a-z+]+:)?\/\//.test(href)) {
                link.href = path_to_root + href;
            }
            // The 'index' page is supposed to alias the first chapter in the book.
            // Check both with and without the '.html' suffix to be robust against pretty URLs
            if (link.href.replace(/\.html$/, '') === current_page.replace(/\.html$/, '')
                || i === 0
                && path_to_root === ''
                && current_page.endsWith('/index.html')) {
                link.classList.add('active');
                let parent = link.parentElement;
                while (parent) {
                    if (parent.tagName === 'LI' && parent.classList.contains('chapter-item')) {
                        parent.classList.add('expanded');
                    }
                    parent = parent.parentElement;
                }
            }
        }
        // Track and set sidebar scroll position
        this.addEventListener('click', e => {
            if (e.target.tagName === 'A') {
                const clientRect = e.target.getBoundingClientRect();
                const sidebarRect = this.getBoundingClientRect();
                sessionStorage.setItem('sidebar-scroll-offset', clientRect.top - sidebarRect.top);
            }
        }, { passive: true });
        const sidebarScrollOffset = sessionStorage.getItem('sidebar-scroll-offset');
        sessionStorage.removeItem('sidebar-scroll-offset');
        if (sidebarScrollOffset !== null) {
            // preserve sidebar scroll position when navigating via links within sidebar
            const activeSection = this.querySelector('.active');
            if (activeSection) {
                const clientRect = activeSection.getBoundingClientRect();
                const sidebarRect = this.getBoundingClientRect();
                const currentOffset = clientRect.top - sidebarRect.top;
                this.scrollTop += currentOffset - parseFloat(sidebarScrollOffset);
            }
        } else {
            // scroll sidebar to current active section when navigating via
            // 'next/previous chapter' buttons
            const activeSection = document.querySelector('#mdbook-sidebar .active');
            if (activeSection) {
                activeSection.scrollIntoView({ block: 'center' });
            }
        }
        // Toggle buttons
        const sidebarAnchorToggles = document.querySelectorAll('.chapter-fold-toggle');
        function toggleSection(ev) {
            ev.currentTarget.parentElement.parentElement.classList.toggle('expanded');
        }
        Array.from(sidebarAnchorToggles).forEach(el => {
            el.addEventListener('click', toggleSection);
        });
    }
}
window.customElements.define('mdbook-sidebar-scrollbox', MDBookSidebarScrollbox);


// ---------------------------------------------------------------------------
// Support for dynamically adding headers to the sidebar.

(function() {
    // This is used to detect which direction the page has scrolled since the
    // last scroll event.
    let lastKnownScrollPosition = 0;
    // This is the threshold in px from the top of the screen where it will
    // consider a header the "current" header when scrolling down.
    const defaultDownThreshold = 150;
    // Same as defaultDownThreshold, except when scrolling up.
    const defaultUpThreshold = 300;
    // The threshold is a virtual horizontal line on the screen where it
    // considers the "current" header to be above the line. The threshold is
    // modified dynamically to handle headers that are near the bottom of the
    // screen, and to slightly offset the behavior when scrolling up vs down.
    let threshold = defaultDownThreshold;
    // This is used to disable updates while scrolling. This is needed when
    // clicking the header in the sidebar, which triggers a scroll event. It
    // is somewhat finicky to detect when the scroll has finished, so this
    // uses a relatively dumb system of disabling scroll updates for a short
    // time after the click.
    let disableScroll = false;
    // Array of header elements on the page.
    let headers;
    // Array of li elements that are initially collapsed headers in the sidebar.
    // I'm not sure why eslint seems to have a false positive here.
    // eslint-disable-next-line prefer-const
    let headerToggles = [];
    // This is a debugging tool for the threshold which you can enable in the console.
    let thresholdDebug = false;

    // Updates the threshold based on the scroll position.
    function updateThreshold() {
        const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
        const windowHeight = window.innerHeight;
        const documentHeight = document.documentElement.scrollHeight;

        // The number of pixels below the viewport, at most documentHeight.
        // This is used to push the threshold down to the bottom of the page
        // as the user scrolls towards the bottom.
        const pixelsBelow = Math.max(0, documentHeight - (scrollTop + windowHeight));
        // The number of pixels above the viewport, at least defaultDownThreshold.
        // Similar to pixelsBelow, this is used to push the threshold back towards
        // the top when reaching the top of the page.
        const pixelsAbove = Math.max(0, defaultDownThreshold - scrollTop);
        // How much the threshold should be offset once it gets close to the
        // bottom of the page.
        const bottomAdd = Math.max(0, windowHeight - pixelsBelow - defaultDownThreshold);
        let adjustedBottomAdd = bottomAdd;

        // Adjusts bottomAdd for a small document. The calculation above
        // assumes the document is at least twice the windowheight in size. If
        // it is less than that, then bottomAdd needs to be shrunk
        // proportional to the difference in size.
        if (documentHeight < windowHeight * 2) {
            const maxPixelsBelow = documentHeight - windowHeight;
            const t = 1 - pixelsBelow / Math.max(1, maxPixelsBelow);
            const clamp = Math.max(0, Math.min(1, t));
            adjustedBottomAdd *= clamp;
        }

        let scrollingDown = true;
        if (scrollTop < lastKnownScrollPosition) {
            scrollingDown = false;
        }

        if (scrollingDown) {
            // When scrolling down, move the threshold up towards the default
            // downwards threshold position. If near the bottom of the page,
            // adjustedBottomAdd will offset the threshold towards the bottom
            // of the page.
            const amountScrolledDown = scrollTop - lastKnownScrollPosition;
            const adjustedDefault = defaultDownThreshold + adjustedBottomAdd;
            threshold = Math.max(adjustedDefault, threshold - amountScrolledDown);
        } else {
            // When scrolling up, move the threshold down towards the default
            // upwards threshold position. If near the bottom of the page,
            // quickly transition the threshold back up where it normally
            // belongs.
            const amountScrolledUp = lastKnownScrollPosition - scrollTop;
            const adjustedDefault = defaultUpThreshold - pixelsAbove
                + Math.max(0, adjustedBottomAdd - defaultDownThreshold);
            threshold = Math.min(adjustedDefault, threshold + amountScrolledUp);
        }

        if (documentHeight <= windowHeight) {
            threshold = 0;
        }

        if (thresholdDebug) {
            const id = 'mdbook-threshold-debug-data';
            let data = document.getElementById(id);
            if (data === null) {
                data = document.createElement('div');
                data.id = id;
                data.style.cssText = `
                    position: fixed;
                    top: 50px;
                    right: 10px;
                    background-color: 0xeeeeee;
                    z-index: 9999;
                    pointer-events: none;
                `;
                document.body.appendChild(data);
            }
            data.innerHTML = `
                <table>
                  <tr><td>documentHeight</td><td>${documentHeight.toFixed(1)}</td></tr>
                  <tr><td>windowHeight</td><td>${windowHeight.toFixed(1)}</td></tr>
                  <tr><td>scrollTop</td><td>${scrollTop.toFixed(1)}</td></tr>
                  <tr><td>pixelsAbove</td><td>${pixelsAbove.toFixed(1)}</td></tr>
                  <tr><td>pixelsBelow</td><td>${pixelsBelow.toFixed(1)}</td></tr>
                  <tr><td>bottomAdd</td><td>${bottomAdd.toFixed(1)}</td></tr>
                  <tr><td>adjustedBottomAdd</td><td>${adjustedBottomAdd.toFixed(1)}</td></tr>
                  <tr><td>scrollingDown</td><td>${scrollingDown}</td></tr>
                  <tr><td>threshold</td><td>${threshold.toFixed(1)}</td></tr>
                </table>
            `;
            drawDebugLine();
        }

        lastKnownScrollPosition = scrollTop;
    }

    function drawDebugLine() {
        if (!document.body) {
            return;
        }
        const id = 'mdbook-threshold-debug-line';
        const existingLine = document.getElementById(id);
        if (existingLine) {
            existingLine.remove();
        }
        const line = document.createElement('div');
        line.id = id;
        line.style.cssText = `
            position: fixed;
            top: ${threshold}px;
            left: 0;
            width: 100vw;
            height: 2px;
            background-color: red;
            z-index: 9999;
            pointer-events: none;
        `;
        document.body.appendChild(line);
    }

    function mdbookEnableThresholdDebug() {
        thresholdDebug = true;
        updateThreshold();
        drawDebugLine();
    }

    window.mdbookEnableThresholdDebug = mdbookEnableThresholdDebug;

    // Updates which headers in the sidebar should be expanded. If the current
    // header is inside a collapsed group, then it, and all its parents should
    // be expanded.
    function updateHeaderExpanded(currentA) {
        // Add expanded to all header-item li ancestors.
        let current = currentA.parentElement;
        while (current) {
            if (current.tagName === 'LI' && current.classList.contains('header-item')) {
                current.classList.add('expanded');
            }
            current = current.parentElement;
        }
    }

    // Updates which header is marked as the "current" header in the sidebar.
    // This is done with a virtual Y threshold, where headers at or below
    // that line will be considered the current one.
    function updateCurrentHeader() {
        if (!headers || !headers.length) {
            return;
        }

        // Reset the classes, which will be rebuilt below.
        const els = document.getElementsByClassName('current-header');
        for (const el of els) {
            el.classList.remove('current-header');
        }
        for (const toggle of headerToggles) {
            toggle.classList.remove('expanded');
        }

        // Find the last header that is above the threshold.
        let lastHeader = null;
        for (const header of headers) {
            const rect = header.getBoundingClientRect();
            if (rect.top <= threshold) {
                lastHeader = header;
            } else {
                break;
            }
        }
        if (lastHeader === null) {
            lastHeader = headers[0];
            const rect = lastHeader.getBoundingClientRect();
            const windowHeight = window.innerHeight;
            if (rect.top >= windowHeight) {
                return;
            }
        }

        // Get the anchor in the summary.
        const href = '#' + lastHeader.id;
        const a = [...document.querySelectorAll('.header-in-summary')]
            .find(element => element.getAttribute('href') === href);
        if (!a) {
            return;
        }

        a.classList.add('current-header');

        updateHeaderExpanded(a);
    }

    // Updates which header is "current" based on the threshold line.
    function reloadCurrentHeader() {
        if (disableScroll) {
            return;
        }
        updateThreshold();
        updateCurrentHeader();
    }


    // When clicking on a header in the sidebar, this adjusts the threshold so
    // that it is located next to the header. This is so that header becomes
    // "current".
    function headerThresholdClick(event) {
        // See disableScroll description why this is done.
        disableScroll = true;
        setTimeout(() => {
            disableScroll = false;
        }, 100);
        // requestAnimationFrame is used to delay the update of the "current"
        // header until after the scroll is done, and the header is in the new
        // position.
        requestAnimationFrame(() => {
            requestAnimationFrame(() => {
                // Closest is needed because if it has child elements like <code>.
                const a = event.target.closest('a');
                const href = a.getAttribute('href');
                const targetId = href.substring(1);
                const targetElement = document.getElementById(targetId);
                if (targetElement) {
                    threshold = targetElement.getBoundingClientRect().bottom;
                    updateCurrentHeader();
                }
            });
        });
    }

    // Takes the nodes from the given head and copies them over to the
    // destination, along with some filtering.
    function filterHeader(source, dest) {
        const clone = source.cloneNode(true);
        clone.querySelectorAll('mark').forEach(mark => {
            mark.replaceWith(...mark.childNodes);
        });
        dest.append(...clone.childNodes);
    }

    // Scans page for headers and adds them to the sidebar.
    document.addEventListener('DOMContentLoaded', function() {
        const activeSection = document.querySelector('#mdbook-sidebar .active');
        if (activeSection === null) {
            return;
        }

        const main = document.getElementsByTagName('main')[0];
        headers = Array.from(main.querySelectorAll('h2, h3, h4, h5, h6'))
            .filter(h => h.id !== '' && h.children.length && h.children[0].tagName === 'A');

        if (headers.length === 0) {
            return;
        }

        // Build a tree of headers in the sidebar.

        const stack = [];

        const firstLevel = parseInt(headers[0].tagName.charAt(1));
        for (let i = 1; i < firstLevel; i++) {
            const ol = document.createElement('ol');
            ol.classList.add('section');
            if (stack.length > 0) {
                stack[stack.length - 1].ol.appendChild(ol);
            }
            stack.push({level: i + 1, ol: ol});
        }

        // The level where it will start folding deeply nested headers.
        const foldLevel = 3;

        for (let i = 0; i < headers.length; i++) {
            const header = headers[i];
            const level = parseInt(header.tagName.charAt(1));

            const currentLevel = stack[stack.length - 1].level;
            if (level > currentLevel) {
                // Begin nesting to this level.
                for (let nextLevel = currentLevel + 1; nextLevel <= level; nextLevel++) {
                    const ol = document.createElement('ol');
                    ol.classList.add('section');
                    const last = stack[stack.length - 1];
                    const lastChild = last.ol.lastChild;
                    // Handle the case where jumping more than one nesting
                    // level, which doesn't have a list item to place this new
                    // list inside of.
                    if (lastChild) {
                        lastChild.appendChild(ol);
                    } else {
                        last.ol.appendChild(ol);
                    }
                    stack.push({level: nextLevel, ol: ol});
                }
            } else if (level < currentLevel) {
                while (stack.length > 1 && stack[stack.length - 1].level > level) {
                    stack.pop();
                }
            }

            const li = document.createElement('li');
            li.classList.add('header-item');
            li.classList.add('expanded');
            if (level < foldLevel) {
                li.classList.add('expanded');
            }
            const span = document.createElement('span');
            span.classList.add('chapter-link-wrapper');
            const a = document.createElement('a');
            span.appendChild(a);
            a.href = '#' + header.id;
            a.classList.add('header-in-summary');
            filterHeader(header.children[0], a);
            a.addEventListener('click', headerThresholdClick);
            const nextHeader = headers[i + 1];
            if (nextHeader !== undefined) {
                const nextLevel = parseInt(nextHeader.tagName.charAt(1));
                if (nextLevel > level && level >= foldLevel) {
                    const toggle = document.createElement('a');
                    toggle.classList.add('chapter-fold-toggle');
                    toggle.classList.add('header-toggle');
                    toggle.addEventListener('click', () => {
                        li.classList.toggle('expanded');
                    });
                    const toggleDiv = document.createElement('div');
                    toggleDiv.textContent = '❱';
                    toggle.appendChild(toggleDiv);
                    span.appendChild(toggle);
                    headerToggles.push(li);
                }
            }
            li.appendChild(span);

            const currentParent = stack[stack.length - 1];
            currentParent.ol.appendChild(li);
        }

        const onThisPage = document.createElement('div');
        onThisPage.classList.add('on-this-page');
        onThisPage.append(stack[0].ol);
        const activeItemSpan = activeSection.parentElement;
        activeItemSpan.after(onThisPage);
    });

    document.addEventListener('DOMContentLoaded', reloadCurrentHeader);
    document.addEventListener('scroll', reloadCurrentHeader, { passive: true });
})();

