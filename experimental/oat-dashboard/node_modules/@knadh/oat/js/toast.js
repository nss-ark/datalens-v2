/**
 * oat - Toast Notifications
 *
 * Usage:
 *   ot.toast('Saved!')
 *   ot.toast('Saved!', 'Your changes have been saved.')
 *   ot.toast('Success', 'Operation completed.', { variant: 'success' })
 *   ot.toast('Error', 'Something went wrong.', { variant: 'danger', placement: 'bottom-center' })
 *
 *   // Custom markup
 *   ot.toastEl(element)
 *   ot.toastEl(element, { duration: 4000, placement: 'bottom-center' })
 *   ot.toastEl(document.querySelector('#my-template'))
 */

const ot = window.ot || (window.ot = {});

const containers = {};
const DEFAULT_DURATION = 4000;
const DEFAULT_PLACEMENT = 'top-right';

function getContainer(placement) {
  if (!containers[placement]) {
    const el = document.createElement('div');
    el.className = 'toast-container';
    el.setAttribute('popover', 'manual');
    el.setAttribute('data-placement', placement);
    document.body.appendChild(el);
    containers[placement] = el;
  }
  return containers[placement];
}

function show(toast, options = {}) {
  const { placement = DEFAULT_PLACEMENT, duration = DEFAULT_DURATION } = options;
  const container = getContainer(placement);

  toast.classList.add('toast');

  let timeout;

  // Pause on hover.
  toast.onmouseenter = () => clearTimeout(timeout);
  toast.onmouseleave = () => {
    if (duration > 0) {
      timeout = setTimeout(() => removeToast(toast, container), duration);
    }
  };

  // Show with animation.
  toast.setAttribute('data-entering', '');
  container.appendChild(toast);
  container.showPopover();

  // Double RAF to compute styles before transition starts.
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      toast.removeAttribute('data-entering');
    });
  });

  if (duration > 0) {
    timeout = setTimeout(() => removeToast(toast, container), duration);
  }

  return toast;
}

// Simple text toast.
ot.toast = function (message, title, options = {}) {
  if (typeof message === 'object' && message !== null) {
    options = message;
    message = '';
  }

  const { variant = '', ...rest } = options;

  // Create toast
  const toast = document.createElement('output');
  toast.className = 'toast';
  toast.setAttribute('role', 'status');
  if (variant) toast.setAttribute('data-variant', variant);

  const titleText = title || (variant[0].toUpperCase() + variant.slice(1));
  const titleEl = document.createElement('h6');
  titleEl.className = 'toast-title';
  if (variant) {
    titleEl.style.color = `var(--${variant})`;
  }
  titleEl.textContent = title || titleText;
  toast.appendChild(titleEl);

  if (message) {
    const msgEl = document.createElement('div');
    msgEl.className = 'toast-message';
    msgEl.textContent = message;
    toast.appendChild(msgEl);
  }

  return show(toast, rest);
};

// Element-based toast.
ot.toastEl = function (el, options = {}) {
  let toast;

  if (el instanceof HTMLTemplateElement) {
    toast = el.content.firstElementChild.cloneNode(true);
  } else if (typeof el === 'string') {
    toast = document.querySelector(el).cloneNode(true);
  } else {
    toast = el.cloneNode(true);
  }

  toast.removeAttribute('id');

  return show(toast, options);
};

function removeToast(toast, container) {
  if (toast.hasAttribute('data-exiting')) {
    return;
  }
  toast.setAttribute('data-exiting', '');

  let done = false;
  const cleanup = () => {
    if (done) {
      return;
    }
    done = true;
    toast.remove();
    if (!container.children.length) {
      container.hidePopover();
    }
  };

  toast.addEventListener('transitionend', cleanup, { once: true });
  setTimeout(cleanup, 200);
}

// Clear all toasts.
ot.toast.clear = function (placement) {
  if (placement && containers[placement]) {
    containers[placement].innerHTML = '';
    containers[placement].hidePopover();
  } else {
    Object.values(containers).forEach(c => {
      c.innerHTML = '';
      c.hidePopover();
    });
  }
};
