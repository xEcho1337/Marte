const App = {
    token() {
        return localStorage.getItem('token');
    },

    isLoggedIn() {
        return !!this.token();
    },

    logout() {
        localStorage.removeItem('token');
        window.location.href = '/login';
    },

    requireAuth() {
        if (!this.isLoggedIn()) {
            window.location.href = '/login';
            return false;
        }
        return true;
    },

    async api(path, options = {}) {
        const headers = options.headers || {};
        const token = this.token();
        if (token) {
            headers['Authorization'] = 'Bearer ' + token;
        }
        if (options.body && !headers['Content-Type']) {
            headers['Content-Type'] = 'application/json';
        }

        const res = await fetch(path, { ...options, headers });
        if (res.status === 401) {
            this.logout();
            return null;
        }
        return res;
    },

    async apiJSON(path, options = {}) {
        const res = await this.api(path, options);
        if (!res || !res.ok) return null;
        return res.json();
    },

    async login(username, password) {
        let res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (res.status === 429) {
            return { ok: false, error: 'Too many attempts, try again later' };
        }

        if (res.status === 401) {
            return { ok: false, error: 'Invalid credentials' };
        }

        try {
            let data = await res.json();

            if (res.ok && data.authorization) {
                localStorage.setItem('token', data.authorization);
                return { ok: true };
            }

            return { ok: false, error: data.error || 'Wrong credentials' };
        } catch (e) {
            return { ok: false, error: 'Server error' };
        }
    },

    async getProfile() {
        return this.apiJSON('/api/profile');
    },

    async getFarm() {
        return this.apiJSON('/api/farm');
    },

    async getFlags(limit = 0, offset = 0) {
        const params = new URLSearchParams();
        if (limit > 0) params.set('limit', limit);
        if (offset > 0) params.set('offset', offset);
        const qs = params.toString();
        return this.apiJSON(`/api/flags${qs ? '?' + qs : ''}`);
    },

    async getServices() {
        return this.apiJSON('/api/get_services');
    }
};
