document.addEventListener('DOMContentLoaded', function() {
    // Upload dialog functionality
    const uploadBtn = document.getElementById('uploadBtn');
    const uploadDialog = document.getElementById('uploadDialog');
    const uploadCloseBtn = uploadDialog.querySelector('.close-btn');
    
    // Preview dialog functionality
    const previewBtn = document.getElementById('previewBtn');
    const previewDialog = document.getElementById('previewDialog');
    const previewCloseBtn = previewDialog.querySelector('.close-btn');

    // Upload dialog event listeners
    uploadBtn.addEventListener('click', () => {
        uploadDialog.showModal();
    });

    uploadCloseBtn.addEventListener('click', () => {
        uploadDialog.close();
    });

    uploadDialog.addEventListener('click', (e) => {
        if (e.target === uploadDialog) {
            uploadDialog.close();
        }
    });

    // Preview dialog event listeners
    previewBtn.addEventListener('click', () => {
        previewDialog.showModal();
    });

    previewCloseBtn.addEventListener('click', () => {
        previewDialog.close();
    });

    previewDialog.addEventListener('click', (e) => {
        if (e.target === previewDialog) {
            previewDialog.close();
        }
    });

    // Preview controls functionality
    const textInput = document.getElementById('previewText');
    const sizeInput = document.getElementById('fontSize');
    const sizeValue = document.getElementById('fontSizeValue');
    const previews = document.querySelectorAll('.preview');

    function updatePreviews() {
        const text = textInput.value || 'The quick brown fox jumps over the lazy dog';
        const size = sizeInput.value + 'px';
        
        previews.forEach(preview => {
            preview.textContent = text;
            preview.style.fontSize = size;
        });
        
        sizeValue.textContent = sizeInput.value + 'px';
    }

    textInput.addEventListener('input', updatePreviews);
    sizeInput.addEventListener('input', updatePreviews);

    updatePreviews();
});