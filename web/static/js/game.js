// Dictionary Game - Client-Side JavaScript

// ===========================
// Toast Notifications
// ===========================

function showToast(message, type = 'info') {
    const container = document.getElementById('toast-container');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;

    container.appendChild(toast);

    // Auto-remove after 3 seconds
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// ===========================
// Timer Countdown
// ===========================

function initializeTimers() {
    const timers = document.querySelectorAll('.timer[data-phase-start]');

    timers.forEach(timer => {
        const phaseStart = new Date(timer.dataset.phaseStart);
        const duration = parseInt(timer.dataset.duration); // seconds

        function updateTimer() {
            const now = new Date();
            const elapsed = Math.floor((now - phaseStart) / 1000);
            const remaining = Math.max(0, duration - elapsed);

            const minutes = Math.floor(remaining / 60);
            const seconds = remaining % 60;

            timer.textContent = `${minutes}:${seconds.toString().padStart(2, '0')}`;

            // Change color when time is running out
            if (remaining <= 10) {
                timer.style.color = 'var(--danger)';
            } else if (remaining <= 30) {
                timer.style.color = 'var(--warning)';
            }

            if (remaining > 0) {
                requestAnimationFrame(updateTimer);
            }
        }

        updateTimer();
    });
}

// Re-initialize timers when phase changes
document.addEventListener('htmx:afterSwap', (event) => {
    if (event.detail.target.id === 'game-container') {
        initializeTimers();
    }
});

// Initialize on page load
window.addEventListener('DOMContentLoaded', initializeTimers);

// ===========================
// Template Helper Functions
// (These would normally be server-side, but including for reference)
// ===========================

// Custom template functions that Go templates don't have

// Helper for template: add numbers
function add(a, b) {
    return a + b;
}

// Helper for template: subtract numbers
function sub(a, b) {
    return a - b;
}

// Helper for template: divide and convert to float
function divf(a, b) {
    return parseFloat(a) / parseFloat(b);
}

// Helper for template: multiply float
function mulf(a, b) {
    return a * b;
}

// Helper for template: find max score
function maxScore(scores) {
    return Math.max(...Object.values(scores));
}

// ===========================
// Form Enhancements
// ===========================

// Auto-save nickname to localStorage
document.addEventListener('DOMContentLoaded', () => {
    // Load saved nickname
    const savedNickname = localStorage.getItem('nickname');
    const nicknameInputs = document.querySelectorAll('input[name="nickname"]');

    if (savedNickname) {
        nicknameInputs.forEach(input => {
            if (!input.value) {
                input.value = savedNickname;
            }
        });
    }

    // Save on change
    nicknameInputs.forEach(input => {
        input.addEventListener('change', (e) => {
            localStorage.setItem('nickname', e.target.value);
        });
    });
});

// ===========================
// HTMX Event Handlers
// ===========================

// Log HTMX errors for debugging
document.addEventListener('htmx:responseError', (event) => {
    console.error('HTMX Error:', event.detail);
    showToast('An error occurred. Please try again.', 'error');
});

// Show loading state
document.addEventListener('htmx:beforeRequest', (event) => {
    const button = event.detail.elt.querySelector('button[type="submit"]');
    if (button) {
        button.dataset.originalText = button.textContent;
        button.textContent = 'Loading...';
        button.disabled = true;
    }
});

// Reset loading state
document.addEventListener('htmx:afterRequest', (event) => {
    const button = event.detail.elt.querySelector('button[type="submit"]');
    if (button && button.dataset.originalText) {
        button.textContent = button.dataset.originalText;
        button.disabled = false;
    }
});

// ===========================
// SSE Connection Management
// ===========================

let sseReconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;

document.addEventListener('htmx:sseError', (event) => {
    console.error('SSE Error:', event.detail);

    sseReconnectAttempts++;

    if (sseReconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        const delay = Math.min(1000 * Math.pow(2, sseReconnectAttempts), 10000);
        showToast(`Connection lost. Reconnecting in ${delay / 1000}s...`, 'error');

        setTimeout(() => {
            // HTMX will auto-reconnect
            console.log('Attempting SSE reconnection...');
        }, delay);
    } else {
        showToast('Connection lost. Please refresh the page.', 'error');
    }
});

document.addEventListener('htmx:sseOpen', (event) => {
    console.log('SSE connection opened');
    sseReconnectAttempts = 0; // Reset counter on successful connection
});

// ===========================
// Utility Functions
// ===========================

// Copy text to clipboard
async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch (err) {
        console.error('Failed to copy:', err);
        return false;
    }
}

// Play sound effect (future enhancement)
function playSound(soundName) {
    // Placeholder for sound effects
    // Could load sounds like: /static/sounds/vote-submitted.mp3
}

// ===========================
// Debug Helpers (development only)
// ===========================

if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    // Debug mode: log all HTMX events
    window.htmxDebug = false; // Set to true to enable

    if (window.htmxDebug) {
        document.body.addEventListener('htmx:*', (event) => {
            console.log('[HTMX]', event.type, event.detail);
        });
    }
}

// Export for use in inline scripts
window.showToast = showToast;
window.copyToClipboard = copyToClipboard;
