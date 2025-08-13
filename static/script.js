document.addEventListener('DOMContentLoaded', function() {
    // Initialize pill groups for truncation
    initializePillGroups();
    
    // Initialize search functionality
    initializeSearch();
    
    // Upload dialog functionality
    const uploadBtn = document.getElementById('uploadBtn');
    const uploadDialog = document.getElementById('uploadDialog');
    const uploadCloseBtn = uploadDialog.querySelector('.close-btn');
    
    // Preview dialog functionality
    const previewBtn = document.getElementById('previewBtn');
    const previewDialog = document.getElementById('previewDialog');
    const previewCloseBtn = previewDialog.querySelector('.close-btn');
    
    // Edit dialog functionality
    const editDialog = document.getElementById('editDialog');
    const editCloseBtn = editDialog.querySelector('.close-btn');

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

    // Edit dialog event listeners
    editCloseBtn.addEventListener('click', () => {
        editDialog.close();
    });

    editDialog.addEventListener('click', (e) => {
        if (e.target === editDialog) {
            editDialog.close();
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

    // Random font styling for h1
    function applyRandomFonts() {
        const h1 = document.querySelector('h1');
        if (!h1) return;

        // Collect all available font names from the page
        const fontItems = document.querySelectorAll('.font-name');
        const fontNames = Array.from(fontItems).map(item => item.textContent.trim());
        
        if (fontNames.length === 0) return;

        // Split h1 text into individual characters and wrap each in a span
        const text = h1.textContent;
        const wrappedText = text.split('').map(char => {
            if (char === ' ') {
                return ' ';
            }
            const randomFont = fontNames[Math.floor(Math.random() * fontNames.length)];
            return `<span style="font-family: '${randomFont}'">${char}</span>`;
        }).join('');
        
        h1.innerHTML = wrappedText;
    }

    // Apply random fonts
    applyRandomFonts();
});

// Initialize pill groups to show only first 3 with +N badge
function initializePillGroups() {
    document.querySelectorAll('.pill-group').forEach(group => {
        const pills = group.querySelectorAll('.pill');
        if (pills.length > 3) {
            // Hide pills beyond the first 3
            for (let i = 3; i < pills.length; i++) {
                pills[i].classList.add('hidden');
            }
            // Add the +N badge
            const morePill = document.createElement('span');
            morePill.className = 'pill more-pill';
            morePill.textContent = `+${pills.length - 3}`;
            morePill.onclick = () => togglePills(morePill);
            group.appendChild(morePill);
        }
    });
}

// Global function to toggle pills visibility
function togglePills(morePill) {
    const pillGroup = morePill.parentElement;
    const allPills = pillGroup.querySelectorAll('.pill:not(.more-pill)');
    const isExpanded = morePill.classList.contains('expanded');
    
    if (isExpanded) {
        // Collapse: hide pills beyond the first 3
        for (let i = 3; i < allPills.length; i++) {
            allPills[i].classList.add('hidden');
        }
        morePill.classList.remove('expanded');
        morePill.textContent = `+${allPills.length - 3}`;
    } else {
        // Expand: show all pills
        allPills.forEach(pill => pill.classList.remove('hidden'));
        morePill.classList.add('expanded');
        morePill.textContent = 'Show less';
    }
}

// Search functionality
function initializeSearch() {
    const searchInput = document.getElementById('searchInput');
    const searchSuggestions = document.getElementById('searchSuggestions');
    const fontItems = document.querySelectorAll('.font-item');
    
    // Collect all unique values for autocomplete
    const autocompleteData = {
        commercialUse: ['Commercial use', 'Personal use only'],
        projects: new Set(),
        tags: new Set()
    };
    
    // Extract projects and tags from all font items
    fontItems.forEach(item => {
        const projectPills = item.querySelectorAll('.project-pill');
        const tagPills = item.querySelectorAll('.tag-pill');
        
        projectPills.forEach(pill => autocompleteData.projects.add(pill.textContent.trim()));
        tagPills.forEach(pill => autocompleteData.tags.add(pill.textContent.trim()));
    });
    
    // Convert sets to arrays
    autocompleteData.projects = Array.from(autocompleteData.projects);
    autocompleteData.tags = Array.from(autocompleteData.tags);
    
    let currentSuggestionIndex = -1;
    
    searchInput.addEventListener('input', function() {
        const query = this.value.toLowerCase().trim();
        
        if (query === '') {
            // Show all fonts when search is empty
            fontItems.forEach(item => item.style.display = 'block');
            searchSuggestions.innerHTML = '';
            searchSuggestions.style.display = 'none';
            return;
        }
        
        // Filter fonts
        fontItems.forEach(item => {
            const fontName = item.querySelector('.font-name').textContent.toLowerCase();
            const fontFilename = item.querySelector('.font-filename').textContent.toLowerCase();
            const source = item.querySelector('.metadata-item').textContent.toLowerCase();
            const commercialUse = item.querySelector('.commercial-use-pill, .personal-use-pill').textContent.toLowerCase();
            
            const projectPills = item.querySelectorAll('.project-pill');
            const projects = Array.from(projectPills).map(pill => pill.textContent.toLowerCase()).join(' ');
            
            const tagPills = item.querySelectorAll('.tag-pill');
            const tags = Array.from(tagPills).map(pill => pill.textContent.toLowerCase()).join(' ');
            
            const searchText = `${fontName} ${fontFilename} ${source} ${commercialUse} ${projects} ${tags}`;
            
            if (searchText.includes(query)) {
                item.style.display = 'block';
            } else {
                item.style.display = 'none';
            }
        });
        
        // Show suggestions
        showSuggestions(query, autocompleteData);
    });
    
    // Handle keyboard navigation for suggestions
    searchInput.addEventListener('keydown', function(e) {
        const suggestions = searchSuggestions.querySelectorAll('.suggestion-item');
        
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            currentSuggestionIndex = Math.min(currentSuggestionIndex + 1, suggestions.length - 1);
            updateSuggestionHighlight(suggestions);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            currentSuggestionIndex = Math.max(currentSuggestionIndex - 1, -1);
            updateSuggestionHighlight(suggestions);
        } else if (e.key === 'Enter' && currentSuggestionIndex >= 0) {
            e.preventDefault();
            suggestions[currentSuggestionIndex].click();
        } else if (e.key === 'Escape') {
            searchSuggestions.style.display = 'none';
            currentSuggestionIndex = -1;
        }
    });
    
    // Hide suggestions when clicking outside
    document.addEventListener('click', function(e) {
        if (!searchInput.contains(e.target) && !searchSuggestions.contains(e.target)) {
            searchSuggestions.style.display = 'none';
            currentSuggestionIndex = -1;
        }
    });
    
    function showSuggestions(query, data) {
        const suggestions = [];
        
        // Add commercial use suggestions
        data.commercialUse.forEach(item => {
            if (item.toLowerCase().includes(query)) {
                suggestions.push({ text: item, type: 'commercial' });
            }
        });
        
        // Add project suggestions
        data.projects.forEach(item => {
            if (item.toLowerCase().includes(query)) {
                suggestions.push({ text: item, type: 'project' });
            }
        });
        
        // Add tag suggestions
        data.tags.forEach(item => {
            if (item.toLowerCase().includes(query)) {
                suggestions.push({ text: item, type: 'tag' });
            }
        });
        
        if (suggestions.length > 0) {
            searchSuggestions.innerHTML = suggestions.slice(0, 8).map((suggestion, index) => 
                `<div class="suggestion-item" data-text="${suggestion.text}">${suggestion.text} <span class="suggestion-type">${suggestion.type}</span></div>`
            ).join('');
            
            searchSuggestions.style.display = 'block';
            currentSuggestionIndex = -1;
            
            // Add click handlers to suggestions
            searchSuggestions.querySelectorAll('.suggestion-item').forEach(item => {
                item.addEventListener('click', function() {
                    searchInput.value = this.dataset.text;
                    searchInput.dispatchEvent(new Event('input'));
                    searchSuggestions.style.display = 'none';
                    currentSuggestionIndex = -1;
                });
            });
        } else {
            searchSuggestions.style.display = 'none';
            currentSuggestionIndex = -1;
        }
    }
    
    function updateSuggestionHighlight(suggestions) {
        suggestions.forEach((item, index) => {
            if (index === currentSuggestionIndex) {
                item.classList.add('highlighted');
            } else {
                item.classList.remove('highlighted');
            }
        });
    }
}

// Global function to open edit modal
function openEditModal(filename, name, source, commercialUse, projects, tags) {
    const editDialog = document.getElementById('editDialog');
    
    // Populate form fields
    document.getElementById('editFilename').value = filename;
    document.getElementById('editFontName').value = name;
    document.getElementById('editSource').value = source;
    
    // Set commercial use radio button
    if (commercialUse) {
        document.getElementById('editCommercialUseAllowed').checked = true;
    } else {
        document.getElementById('editCommercialUseNotAllowed').checked = true;
    }
    
    // Clean up projects and tags (remove trailing comma)
    const cleanProjects = projects.replace(/,$/, '');
    const cleanTags = tags.replace(/,$/, '');
    
    document.getElementById('editProjects').value = cleanProjects;
    document.getElementById('editTags').value = cleanTags;
    
    editDialog.showModal();
}