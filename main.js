const form = document.getElementById('shorten-form');
const urlInput = document.getElementById('url-input');
const customShortInput = document.getElementById('custom-short-input');
const expiryInput = document.getElementById('expiry-input');
const submitBtn = document.getElementById('submit-btn');

const resultContainer = document.getElementById('result-container');
const shortUrlLink = document.getElementById('short-url');
const copyBtn = document.getElementById('copy-btn');
const copyText = document.getElementById('copy-text');

const errorContainer = document.getElementById('error-container');
const errorMessage = document.getElementById('error-message');

form.addEventListener('submit', async (e) => {
    e.preventDefault();

    // Hide previous results/errors
    resultContainer.classList.add('hidden');
    errorContainer.classList.add('hidden');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Shortening...';

    const longURL = urlInput.value;
    const customShort = customShortInput.value;
    const expiry = parseInt(expiryInput.value, 10);

    try {
        const response = await fetch('/api/v1', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                url: longURL,
                custom_short: customShort,
                expiry: expiry,
            }),
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'An unknown error occurred.');
        }

        shortUrlLink.href = data.custom_short;
        shortUrlLink.textContent = data.custom_short.replace(/^https?:\/\//, '');
        resultContainer.classList.remove('hidden');
        form.reset();

    } catch (error) {
        errorMessage.textContent = `Error: ${error.message}`;
        errorContainer.classList.remove('hidden');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Shorten URL';
    }
});

copyBtn.addEventListener('click', () => {
    navigator.clipboard.writeText(shortUrlLink.href).then(() => {
        copyText.textContent = 'Copied!';
        copyBtn.style.backgroundColor = '#03dac6'; // A success color
        setTimeout(() => {
            copyText.textContent = 'Copy';
            copyBtn.style.backgroundColor = ''; // Revert to original
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy: ', err);
        copyText.textContent = 'Failed';
         setTimeout(() => {
            copyText.textContent = 'Copy';
        }, 2000);
    });
});