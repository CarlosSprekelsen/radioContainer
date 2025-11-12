/**
 * RCC Web UI - OpenAPI v1 + SSE v1 Client
 * Implements Gate A/B/C/D conformance with Architecture v1, CB-TIMING v0.3
 */

class RCCClient {
    constructor() {
        this.baseURL = '';
        this.config = null;
        this.eventSource = null;
        this.lastEventId = localStorage.getItem('lastEventId') || '';
        this.activeRadioId = null;
        this.retryTimeouts = new Map();
        this.pendingOperations = new Map();
        
        this.init();
    }
    
    async init() {
        try {
            await this.loadConfig();
            await this.loadRadios();
            this.connectTelemetry();
            this.setupEventHandlers();
        } catch (error) {
            this.showToast('Failed to initialize: ' + error.message, 'error');
            console.error('Initialization error:', error);
        }
    }
    
    async loadConfig() {
        try {
            const response = await fetch('/config.json');
            this.config = await response.json();
            this.baseURL = this.config.rccBaseUrl || 'http://localhost:8080';
        } catch (error) {
            console.warn('Failed to load config, using defaults:', error);
            this.config = {
                timing: {
                    cmdTimeoutsSec: { setPower: 10, setChannel: 30, selectRadio: 5, getState: 5 },
                    retry: { busyBaseMs: 1000, unavailableBaseMs: 2000, jitterMs: 200 }
                }
            };
            this.baseURL = 'http://localhost:8080';
        }
    }
    
    async apiCall(endpoint, options = {}) {
        const startTime = Date.now();
        const correlationId = this.generateCorrelationId();
        
        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                'X-Correlation-ID': correlationId
            }
        };
        
        const mergedOptions = { ...defaultOptions, ...options };
        if (mergedOptions.headers) {
            mergedOptions.headers = { ...defaultOptions.headers, ...mergedOptions.headers };
        }
        
        try {
            const response = await fetch(endpoint, mergedOptions);
            const data = await response.json();
            
            const latency = Date.now() - startTime;
            this.logAudit({
                timestamp: new Date().toISOString(),
                actor: 'ui',
                radioId: this.activeRadioId || '',
                action: options.method || 'GET',
                result: response.ok ? 'success' : 'error',
                latencyMs: latency,
                correlationId: correlationId
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${data.message || 'Unknown error'}`);
            }
            
            return data;
        } catch (error) {
            const latency = Date.now() - startTime;
            this.logAudit({
                timestamp: new Date().toISOString(),
                actor: 'ui',
                radioId: this.activeRadioId || '',
                action: options.method || 'GET',
                result: 'error',
                latencyMs: latency,
                correlationId: correlationId
            });
            throw error;
        }
    }
    
    async loadRadios() {
        try {
            const data = await this.apiCall('/radios');
            
            if (data.result !== 'ok') {
                throw new Error('Invalid response format');
            }
            
            const radioSelect = document.getElementById('radioSelect');
            radioSelect.innerHTML = '';
            
            if (data.data.radios && data.data.radios.length > 0) {
                data.data.radios.forEach(radio => {
                    const option = document.createElement('option');
                    option.value = radio.id;
                    option.textContent = radio.id;
                    if (radio.id === data.data.activeRadioId) {
                        option.selected = true;
                        this.activeRadioId = radio.id;
                    }
                    radioSelect.appendChild(option);
                });
            } else {
                const option = document.createElement('option');
                option.value = '';
                option.textContent = 'No radios available';
                radioSelect.appendChild(option);
            }
        } catch (error) {
            this.showToast('Failed to load radios: ' + error.message, 'error');
            console.error('Load radios error:', error);
        }
    }
    
    async selectRadio(radioId) {
        if (!radioId) return;
        
        try {
            const data = await this.apiCall('/radios/select', {
                method: 'POST',
                body: JSON.stringify({ id: radioId })
            });
            
            if (data.result === 'ok') {
                this.activeRadioId = radioId;
                this.showToast(`Selected radio: ${radioId}`, 'success');
                await this.loadRadioState();
            } else {
                throw new Error(data.message || 'Failed to select radio');
            }
        } catch (error) {
            this.showToast('Failed to select radio: ' + error.message, 'error');
            console.error('Select radio error:', error);
        }
    }
    
    async loadRadioState() {
        if (!this.activeRadioId) return;
        
        try {
            const [powerData, channelData] = await Promise.all([
                this.apiCall(`/radios/${this.activeRadioId}/power`),
                this.apiCall(`/radios/${this.activeRadioId}/channel`)
            ]);
            
            if (powerData.result === 'ok' && powerData.data) {
                document.getElementById('currentPower').textContent = powerData.data.powerDbm;
                document.getElementById('powerSlider').value = powerData.data.powerDbm;
            }
            
            if (channelData.result === 'ok' && channelData.data) {
                const freq = channelData.data.frequencyMhz;
                const index = channelData.data.channelIndex;
                
                // Display frequency as read-only information
                document.getElementById('currentChannel').textContent = `${freq} MHz`;
                document.getElementById('currentChannelIndex').textContent = index ? `Ch ${index}` : 'Custom';
                
                // Update channel buttons with abstract numbers (1,2,3...)
                this.updateChannelButtons(channelData.data.capabilities);
            }
        } catch (error) {
            console.error('Load radio state error:', error);
        }
    }
    
    async setPower(powerDbm) {
        if (!this.activeRadioId) {
            this.showToast('No radio selected', 'warning');
            return;
        }
        
        const operationKey = `setPower-${this.activeRadioId}`;
        if (this.pendingOperations.has(operationKey)) {
            return; // Prevent duplicate operations
        }
        
        this.pendingOperations.set(operationKey, true);
        
        try {
            const data = await this.apiCall(`/radios/${this.activeRadioId}/power`, {
                method: 'POST',
                body: JSON.stringify({ powerDbm: parseInt(powerDbm) })
            });
            
            if (data.result === 'ok') {
                document.getElementById('currentPower').textContent = powerDbm;
                this.showToast(`Power set to ${powerDbm} dBm`, 'success');
            } else {
                throw new Error(data.message || 'Failed to set power');
            }
        } catch (error) {
            this.handleApiError(error, 'setPower', () => this.setPower(powerDbm));
        } finally {
            this.pendingOperations.delete(operationKey);
        }
    }
    
    async setChannel(channelIndex) {
        if (!this.activeRadioId) {
            this.showToast('No radio selected', 'warning');
            return;
        }
        
        if (!channelIndex) {
            this.showToast('Channel index required', 'warning');
            return;
        }
        
        const operationKey = `setChannel-${this.activeRadioId}`;
        if (this.pendingOperations.has(operationKey)) {
            return;
        }
        
        this.pendingOperations.set(operationKey, true);
        
        try {
            const payload = { channelIndex: parseInt(channelIndex) };
            
            const data = await this.apiCall(`/radios/${this.activeRadioId}/channel`, {
                method: 'POST',
                body: JSON.stringify(payload)
            });
            
            if (data.result === 'ok') {
                this.showToast(`Channel set to ${channelIndex}`, 'success');
                await this.loadRadioState(); // Refresh display
            } else {
                throw new Error(data.message || 'Failed to set channel');
            }
        } catch (error) {
            this.handleApiError(error, 'setChannel', () => this.setChannel(channelIndex));
        } finally {
            this.pendingOperations.delete(operationKey);
        }
    }
    
    handleApiError(error, operation, retryFn) {
        const errorMessage = error.message || 'Unknown error';
        
        if (errorMessage.includes('BUSY') || errorMessage.includes('UNAVAILABLE')) {
            const baseDelay = errorMessage.includes('BUSY') ? 
                this.config.timing.retry.busyBaseMs : 
                this.config.timing.retry.unavailableBaseMs;
            const jitter = Math.random() * this.config.timing.retry.jitterMs;
            const delay = baseDelay + jitter;
            
            this.showToast(`${operation} failed, retrying in ${Math.round(delay/1000)}s...`, 'warning');
            
            const timeoutId = setTimeout(() => {
                this.retryTimeouts.delete(operation);
                retryFn();
            }, delay);
            
            this.retryTimeouts.set(operation, timeoutId);
        } else {
            this.showToast(`${operation} failed: ${errorMessage}`, 'error');
        }
    }
    
    connectTelemetry() {
        if (this.eventSource) {
            this.eventSource.close();
        }
        
        const url = `/telemetry${this.lastEventId ? '?lastEventId=' + encodeURIComponent(this.lastEventId) : ''}`;
        this.eventSource = new EventSource(url);
        
        this.eventSource.onopen = () => {
            this.addTelemetryLog('Connected to telemetry stream', 'ready');
        };
        
        this.eventSource.onerror = (error) => {
            console.error('SSE error:', error);
            this.addTelemetryLog('Telemetry connection error, reconnecting...', 'fault');
            setTimeout(() => this.connectTelemetry(), 5000);
        };
        
        // Handle specific event types per SSE v1 spec
        this.eventSource.addEventListener('ready', (event) => {
            this.addTelemetryLog('System ready', 'ready');
        });
        
        this.eventSource.addEventListener('state', (event) => {
            try {
                const data = JSON.parse(event.data);
                this.updateRadioStatus(data.state);
                this.addTelemetryLog(`State: ${data.state}`, 'state');
            } catch (error) {
                console.error('Failed to parse state event:', error);
            }
        });
        
        this.eventSource.addEventListener('powerChanged', (event) => {
            try {
                const data = JSON.parse(event.data);
                document.getElementById('currentPower').textContent = data.powerDbm;
                document.getElementById('powerSlider').value = data.powerDbm;
                this.addTelemetryLog(`Power changed → ${data.powerDbm} dBm`, 'powerChanged');
            } catch (error) {
                console.error('Failed to parse powerChanged event:', error);
            }
        });
        
        this.eventSource.addEventListener('channelChanged', (event) => {
            try {
                const data = JSON.parse(event.data);
                const freq = data.frequencyMhz;
                const index = data.channelIndex;
                document.getElementById('currentChannel').textContent = `${freq} MHz`;
                document.getElementById('currentChannelIndex').textContent = index ? `Ch ${index}` : 'Custom';
                this.addTelemetryLog(`Channel changed → Ch ${index || 'Custom'} (${freq} MHz)`, 'channelChanged');
            } catch (error) {
                console.error('Failed to parse channelChanged event:', error);
            }
        });
        
        this.eventSource.addEventListener('fault', (event) => {
            try {
                const data = JSON.parse(event.data);
                this.addTelemetryLog(`Fault: ${data.message}`, 'fault');
                this.showToast(`Fault: ${data.message}`, 'error');
            } catch (error) {
                console.error('Failed to parse fault event:', error);
            }
        });
        
        this.eventSource.addEventListener('heartbeat', (event) => {
            this.addTelemetryLog('Heartbeat', 'heartbeat');
        });
        
        // Store last event ID for resume
        this.eventSource.addEventListener('message', (event) => {
            if (event.lastEventId) {
                this.lastEventId = event.lastEventId;
                localStorage.setItem('lastEventId', this.lastEventId);
            }
        });
    }
    
    updateRadioStatus(state) {
        const statusElement = document.getElementById('radioStatus');
        statusElement.className = `status-indicator ${state}`;
        statusElement.textContent = `● ${state}`;
    }
    
    updateChannelButtons(capabilities) {
        const container = document.getElementById('channelButtons');
        container.innerHTML = '';
        
        if (capabilities && capabilities.channels) {
            capabilities.channels.forEach((channel, index) => {
                const button = document.createElement('button');
                button.className = 'channel-btn';
                button.textContent = index + 1; // 1-based indexing per Gate D
                button.onclick = () => this.setChannel(index + 1);
                container.appendChild(button);
            });
        }
    }
    
    addTelemetryLog(message, type = '') {
        const log = document.getElementById('telemetryLog');
        const entry = document.createElement('div');
        entry.className = `log-entry ${type}`;
        entry.textContent = `${new Date().toLocaleTimeString()} ${message}`;
        log.appendChild(entry);
        
        // Keep only last 50 entries
        while (log.children.length > 50) {
            log.removeChild(log.firstChild);
        }
        
        log.scrollTop = log.scrollHeight;
    }
    
    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.textContent = message;
        container.appendChild(toast);
        
        setTimeout(() => {
            toast.remove();
        }, 5000);
    }
    
    logAudit(entry) {
        console.log('AUDIT:', entry);
        
        // Send to server audit endpoint
        fetch('/audit', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(entry)
        }).catch(error => console.error('Failed to send audit log:', error));
    }
    
    generateCorrelationId() {
        return 'ui-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }
    
    setupEventHandlers() {
        // Radio selection
        document.getElementById('radioSelect').addEventListener('change', (e) => {
            this.selectRadio(e.target.value);
        });
        
        // Power control
        document.getElementById('applyPower').addEventListener('click', () => {
            const power = document.getElementById('powerSlider').value;
            this.setPower(power);
        });
        
        // Channel control - only abstract channel numbers (1,2,3...)
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    window.rccClient = new RCCClient();
});
