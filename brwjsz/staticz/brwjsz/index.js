(function() {
    if (window.__brwjsz_initialized) {
        return;
    }
    window.__brwjsz_initialized = true;

    const refs = new Map();
    window.__refs = refs;

    function getTabId() {
        const params = new URLSearchParams(window.location.search);
        return params.get('tab') || '0';
    }

    function log(msg) {
        try {
            fetch('/api/brwjs/log?msg=' + encodeURIComponent('[Tab ' + getTabId() + '] ' + msg)).catch(() => {});
        } catch (_) {}
    }

    function resolveArg(arg) {
        if (arg !== null && typeof arg === 'object') {
            if (arg.$ref && refs.has(arg.$ref)) {
                return refs.get(arg.$ref);
            }
            if (Array.isArray(arg)) {
                return arg.map(resolveArg);
            }
        }
        return arg;
    }

    async function sendReply(reply) {
        try {
            await fetch('/api/brwjs/reply', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(reply)
            });
        } catch (e) {
            console.error('Failed to send reply to /api/brwjs/reply:', e);
        }
    }

    async function handleCommand(cmd) {
        const { id, action, fn, ref, args, value } = cmd;
        log('handling cmd: ' + id + ' action: ' + action);
        try {
            switch (action) {
                case 'proc': {
                    const resolvedArgs = (args || []).map(resolveArg);
                    let target;
                    try {
                        target = (0, eval)('(' + fn + ')');
                    } catch (_) {
                        target = (0, eval)(fn);
                    }
                    if (typeof target === 'function') {
                        const res = target.apply(window, resolvedArgs);
                        if (res && typeof res.then === 'function') {
                            await res;
                        }
                    }
                    await sendReply({ id: id, ok: true });
                    break;
                }
                case 'eval': {
                    let res;
                    try {
                        res = (0, eval)('(' + fn + ')');
                    } catch (_) {
                        res = (0, eval)(fn);
                    }
                    if (res && typeof res.then === 'function') {
                        res = await res;
                    }
                    await sendReply({ id: id, ok: true, result: res });
                    break;
                }
                case 'func': {
                    const resolvedArgs = (args || []).map(resolveArg);
                    let target;
                    try {
                        target = (0, eval)('(' + fn + ')');
                    } catch (_) {
                        target = (0, eval)(fn);
                    }
                    let res;
                    if (typeof target === 'function') {
                        res = target.apply(window, resolvedArgs);
                    } else {
                        res = target;
                    }
                    if (res && typeof res.then === 'function') {
                        res = await res;
                    }
                    refs.set(ref, res);
                    await sendReply({ id: id, ok: true, ref: ref });
                    break;
                }
                case 'set': {
                    refs.set(ref, value);
                    await sendReply({ id: id, ok: true });
                    break;
                }
                case 'get': {
                    if (!refs.has(ref)) {
                        throw new Error('Reference not found: ' + ref);
                    }
                    const val = refs.get(ref);
                    await sendReply({ id: id, ok: true, result: val });
                    break;
                }
                case 'delete': {
                    refs.delete(ref);
                    await sendReply({ id: id, ok: true });
                    break;
                }
                default:
                    throw new Error('Unknown command action: ' + action);
            }
        } catch (err) {
            const errMsg = (err && (err.stack || err.message)) ? (err.stack || err.message) : String(err);
            log('cmd error: ' + id + ' err: ' + errMsg);
            await sendReply({ id: id, ok: false, error: errMsg });
        }
    }

    function init() {
        const tab = getTabId();
        const params = new URLSearchParams(window.location.search);
        const count = parseInt(params.get('count') || '1', 10);
        log('init tab=' + tab + ' count=' + count);

        if (tab === '0' && count > 1) {
            for (let i = 1; i < count; i++) {
                log('tab 0 opening tab ' + i);
                window.open('controller.html?tab=' + i, '_blank');
            }
        }

        const es = new EventSource('/api/brwjs/events?tab=' + encodeURIComponent(tab));

        es.addEventListener('open', () => {
            log('SSE open');
        });

        es.addEventListener('cmd', (event) => {
            log('got cmd event: ' + event.data);
            try {
                const cmd = JSON.parse(event.data);
                handleCommand(cmd);
            } catch (err) {
                log('error parsing cmd event: ' + err);
            }
        });

        es.addEventListener('error', (e) => {
            log('SSE error event');
        });
    }

    window.brwjsz = {
        init: init,
        refs: refs
    };

    if (document.readyState !== 'loading') {
        init();
    } else {
        document.addEventListener('DOMContentLoaded', init);
    }
})();
