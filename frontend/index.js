const express = require('express');
const fetch = (...args) => import('node-fetch').then(({default: fetch}) => fetch(...args));

const app = express();

const GATEWAY_BASE_URL = process.env.GATEWAY_BASE_URL || 'http://gateway:8080';

app.use(express.json());
app.use(express.static('public'));

function proxy(basePath) {
    return async (req, res) => {
        const suffix = req.originalUrl.replace(/^\/api\/[a-z-]+/, '');
        const url = `${GATEWAY_BASE_URL}${basePath}${suffix}`;

        try {
            const init = {
                method: req.method,
                headers: {
                    'Accept': req.headers['accept'] || '*/*',
                    'Content-Type': req.headers['content-type'] || undefined,
                },
            };

            if (!['GET', 'HEAD'].includes(req.method)) {
                init.body = JSON.stringify(req.body ?? {});
                init.headers['Content-Type'] = 'application/json';
            }

            const response = await fetch(url, init);
            const contentType = response.headers.get('content-type') || 'application/json';

            res.status(response.status);
            res.set('Content-Type', contentType);
            const buffer = await response.arrayBuffer();
            res.send(Buffer.from(buffer));
        } catch (err) {
            console.error('[frontend] Ошибка прокси:', err);
            res.status(500).json({error: 'proxy_error', message: err.message});
        }
    };
}

app.all('/api/orders*', proxy('/orders'));
app.all('/api/accounts*', proxy('/accounts'));

app.all('/api/health', proxy('/health'));
app.get('/health', (req, res) => res.status(200).send('ok'));

const port = process.env.PORT || 8080;
app.listen(port, () => {
    console.log(`Frontend server listening on port ${port}`);
});
