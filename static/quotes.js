// Quotes module (the quotes module): the random quote block and the
// quote's form, loaded after app.js (attached by its templates, {{attach}}).

// ===== Random quote =====
// randomQuote - the block's section (#random-quote-section) loads a quote
// from /api/quote/random once it nears the viewport (right away without
// IntersectionObserver); no quote, or an error, hides it. A behavior
// (Kosatka.behaviors), so the block works in a lazy placement too.
Kosatka.behaviors.randomQuote = {
    attach: function (context) {
        Kosatka.once('random-quote', '#random-quote-section', context).forEach(function (section) {
            if (!('IntersectionObserver' in window)) {
                loadRandomQuote(section);
                return;
            }
            const observer = new IntersectionObserver((entries) => {
                if (entries.some((entry) => entry.isIntersecting)) {
                    observer.disconnect();
                    loadRandomQuote(section);
                }
            }, { threshold: 0.1, rootMargin: '100px' });
            observer.observe(section);
        });
    }
};

async function loadRandomQuote(section) {
    const loadingElement = section.querySelector('.quote-loading');
    const contentElement = section.querySelector('.quote-content');
    if (!contentElement) {
        return;
    }
    if (loadingElement) {
        loadingElement.style.display = 'block';
    }
    try {
        const response = await fetch('/api/quote/random');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        if (!data.found) {
            section.classList.add('hidden');
            return;
        }
        contentElement.innerHTML = data.html;
        section.classList.add('quote-loaded');
        Kosatka.attachBehaviors(contentElement);
    } catch (error) {
        console.error('Error loading random quote:', error);
        section.classList.add('hidden');
    } finally {
        if (loadingElement) {
            loadingElement.style.display = 'none';
        }
    }
}

// ===== Quote form =====
// quoteForm - the quote's form (#quote-form, quotes/form.html): its author
// is a Tagify picker (one; an existing author by /api/quotes/authors/search,
// or a new name), written into author_id or author_name; the text is
// required.
Kosatka.behaviors.quoteForm = {
    attach: function (context) {
        Kosatka.once('quote-form', '#quote-form', context).forEach(function (form) {
            if (typeof Tagify === 'undefined') return;
            var authorDisplay = form.querySelector('#author-display');
            var authorIdInput = form.querySelector('#author_id');
            var authorNameInput = form.querySelector('#author_name');

            var initialAuthor = null;
            try {
                var data = authorDisplay.getAttribute('data-initial-author');
                if (data && data !== 'null') {
                    initialAuthor = JSON.parse(data);
                }
            } catch (e) {
                console.warn('Failed to parse initial author:', e);
            }

            var tagify = new Tagify(authorDisplay, {
                tagTextProp: 'value',
                enforceWhitelist: false,
                maxTags: 1,
                dropdown: {
                    enabled: 1,
                    maxItems: 10,
                    closeOnSelect: true,
                    highlightFirst: true
                },
                whitelist: []
            });

            if (initialAuthor && initialAuthor.id) {
                tagify.addTags([initialAuthor]);
                authorIdInput.value = initialAuthor.id;
            }

            tagify.on('add', function(e) {
                updateAuthorInputs();
            });
            tagify.on('remove', function() {
                authorIdInput.value = '';
                authorNameInput.value = '';
            });
            tagify.on('input', function(e) {
                var value = e.detail.value;
                if (!value || value.length < 2) {
                    tagify.dropdown.hide();
                    return;
                }
                fetch('/api/quotes/authors/search?q=' + encodeURIComponent(value))
                    .then(function(r) {
                        if (r.redirected && r.url.indexOf('/login') >= 0) {
                            window.location.href = '/login';
                            return [];
                        }
                        return r.json();
                    })
                    .then(function(data) {
                        if (data && data.suggestions) {
                            var suggestions = data.suggestions.map(function(item) {
                                return {id: item.id, value: item.name};
                            });
                            tagify.settings.whitelist = suggestions;
                            tagify.dropdown.show(value);
                        }
                    })
                    .catch(function(err) {
                        console.error('Author search error:', err);
                    });
            });

            function updateAuthorInputs() {
                var tags = tagify.value;
                if (tags.length === 0) {
                    authorIdInput.value = '';
                    authorNameInput.value = '';
                    return;
                }
                var tag = tags[0];
                if (tag.id && tag.id > 0) {
                    authorIdInput.value = tag.id;
                    authorNameInput.value = '';
                } else {
                    authorIdInput.value = '';
                    authorNameInput.value = tag.value || '';
                }
            }

            form.addEventListener('submit', function(e) {
                updateAuthorInputs();
                if (window.tinymce) tinymce.triggerSave();
                var textField = form.querySelector('#text');
                if (!textField.value.trim()) {
                    e.preventDefault();
                    Kosatka.alert(Kosatka.t('Quote text is required'));
                }
            });
        });
    }
};
