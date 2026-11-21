(function() {
    window.brwqunit = {
        async run(testUrlsOrOpts, runOptions) {
            let testUrls = testUrlsOrOpts;
            let opts = runOptions || {};
            if (testUrlsOrOpts && !Array.isArray(testUrlsOrOpts) && typeof testUrlsOrOpts === 'object') {
                testUrls = testUrlsOrOpts.testUrls || testUrlsOrOpts.urls || [];
                opts = testUrlsOrOpts;
            }
            if (typeof testUrls === 'string') {
                testUrls = [testUrls];
            } else if (!Array.isArray(testUrls)) {
                testUrls = [];
            }

            return new Promise((resolve, reject) => {
                const iframe = document.createElement('iframe');
                iframe.style.width = '100%';
                iframe.style.height = '600px';
                iframe.style.border = 'none';
                document.body.appendChild(iframe);

                let importMapTag = '';
                if (opts.importMap) {
                    let im = opts.importMap;
                    if (typeof im === 'string') {
                        try {
                            im = JSON.parse(im);
                        } catch (_) {}
                    }
                    if (im && typeof im === 'object') {
                        if (!im.imports && !im.scopes) {
                            im = { imports: im };
                        }
                        importMapTag = '<script type="importmap">' + JSON.stringify(im) + '</' + 'script>';
                    }
                }

                const doc = iframe.contentDocument || iframe.contentWindow.document;
                doc.open();
                doc.write('<!DOCTYPE html>' +
                    '<html>' +
                    '<head>' +
                    '<meta charset="utf-8">' +
                    importMapTag +
                    '<link rel="stylesheet" href="/api/brwjs/f/qunit/qunit.css">' +
                    '<script>window.QUnit = { config: { autostart: false } };</' + 'script>' +
                    '<script src="/api/brwjs/f/qunit/qunit.js"></' + 'script>' +
                    '</head>' +
                    '<body>' +
                    '<div id="qunit"></div>' +
                    '<div id="qunit-fixture"></div>' +
                    '</body>' +
                    '</html>');
                doc.close();

                iframe.onload = async () => {
                    try {
                        const win = iframe.contentWindow;
                        const QUnit = win.QUnit;
                        if (!QUnit) {
                            throw new Error('QUnit failed to load in test iframe');
                        }

                        const tests = [];
                        QUnit.on('testEnd', (details) => {
                            const errors = (details.errors || []).map(e => ({
                                message: e.message || String(e),
                                passed: e.passed,
                                actual: e.actual !== undefined ? String(e.actual) : undefined,
                                expected: e.expected !== undefined ? String(e.expected) : undefined,
                                stack: e.stack
                            }));
                            tests.push({
                                name: details.name,
                                suiteName: details.suiteName || '',
                                status: details.status,
                                runtime: details.runtime,
                                errors: errors
                            });
                        });

                        for (const url of testUrls) {
                            const fullUrl = new URL(url, window.location.href).href;
                            await new Promise((resScript, rejScript) => {
                                const script = doc.createElement('script');
                                script.src = fullUrl;
                                if (opts.module || url.endsWith('.mjs')) {
                                    script.type = 'module';
                                }
                                script.onload = resScript;
                                script.onerror = () => rejScript(new Error('Failed to load test script: ' + fullUrl));
                                doc.body.appendChild(script);
                            });
                        }

                        QUnit.on('runEnd', (suite) => {
                            const result = {
                                status: suite.status,
                                total: suite.testCounts.total,
                                passed: suite.testCounts.passed,
                                failed: suite.testCounts.failed,
                                skipped: suite.testCounts.skipped,
                                runtime: suite.runtime,
                                tests: tests
                            };
                            iframe.remove();
                            resolve(result);
                        });

                        QUnit.start();
                    } catch (err) {
                        iframe.remove();
                        reject(err);
                    }
                };
            });
        }
    };
})();
