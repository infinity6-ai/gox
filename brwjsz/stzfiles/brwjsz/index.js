(function() {
    if (window.__brwjsz_initialized) {
        return;
    }
    window.__brwjsz_initialized = true;

    const refs = new Map();
    window.__refs = refs;

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

    function evalTarget(fn) {
        try {
            return (0, eval)('(' + fn + ')');
        } catch (_) {
            return (0, eval)(fn);
        }
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
        try {
            switch (action) {
                case 'proc': {
                    const resolvedArgs = (args || []).map(resolveArg);
                    const target = evalTarget(fn);
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
                    let res = evalTarget(fn);
                    if (res && typeof res.then === 'function') {
                        res = await res;
                    }
                    await sendReply({ id: id, ok: true, result: res });
                    break;
                }
                case 'call': {
                    const resolvedArgs = (args || []).map(resolveArg);
                    const target = evalTarget(fn);
                    let res = (typeof target === 'function') ? target.apply(window, resolvedArgs) : target;
                    if (res && typeof res.then === 'function') {
                        res = await res;
                    }
                    await sendReply({ id: id, ok: true, result: res });
                    break;
                }
                case 'func': {
                    const resolvedArgs = (args || []).map(resolveArg);
                    const target = evalTarget(fn);
                    let res = (typeof target === 'function') ? target.apply(window, resolvedArgs) : target;
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
            await sendReply({ id: id, ok: false, error: errMsg });
        }
    }

    function init() {
        const es = new EventSource('/api/brwjs/events');

        es.addEventListener('open', () => {
            console.log('brwjsz controller connected');
        });

        es.addEventListener('cmd', (event) => {
            try {
                const cmd = JSON.parse(event.data);
                handleCommand(cmd);
            } catch (err) {
                console.error('Error parsing command payload:', err);
            }
        });

        es.addEventListener('error', (e) => {
            console.warn('brwjsz SSE connection status changed:', e);
        });
    }

    function refDump(targetRefs) {
        const summary = {};
        const isTargeted = Array.isArray(targetRefs) && targetRefs.length > 0;
        const entries = [];

        if (isTargeted) {
            for (const item of targetRefs) {
                const id = String(item);
                if (refs.has(id)) {
                    entries.push([id, refs.get(id)]);
                } else {
                    entries.push([id, undefined]);
                }
            }
        } else {
            for (const [id, val] of refs.entries()) {
                entries.push([id, val]);
            }
        }

        console.log(`[brwjsz] RefDump (${entries.length} refs):`);
        for (const [id, val] of entries) {
            let desc;
            if (val === undefined) {
                desc = '<not found>';
            } else if (val && val.constructor && val.constructor.name) {
                desc = val.constructor.name;
            } else if (val !== null && val !== undefined) {
                desc = String(val);
            } else {
                desc = typeof val;
            }
            console.log(`  - [${id}] (${desc}):`, val);
            summary[id] = desc;
        }
        return summary;
    }

    window.brwjsz = {
        init: init,
        refs: refs,
        refDump: refDump
    };

    if (document.readyState !== 'loading') {
        init();
    } else {
        document.addEventListener('DOMContentLoaded', init);
    }
})();
