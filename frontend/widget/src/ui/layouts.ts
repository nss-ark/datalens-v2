export const renderMainContainer = (layout: string): HTMLElement => {
    const container = document.createElement('div');
    container.className = `dl-consent-container dl-layout-${layout.toLowerCase().replace('_', '-')}`;
    return container;
};

export const renderBackdrop = (): HTMLElement => {
    const backdrop = document.createElement('div');
    backdrop.className = 'dl-backdrop';
    return backdrop;
};

export const renderHeader = (titleText: string): HTMLElement => {
    const header = document.createElement('div');
    header.className = 'dl-consent-header';

    const title = document.createElement('h3');
    title.className = 'dl-consent-title';
    title.innerText = titleText;

    header.appendChild(title);
    return header;
};

export const renderDescription = (text: string): HTMLElement => {
    const p = document.createElement('p');
    p.className = 'dl-consent-description';
    p.innerText = text;
    return p;
};

export const renderActions = (): HTMLElement => {
    const div = document.createElement('div');
    div.className = 'dl-consent-actions';
    return div;
};

export const renderButton = (text: string, variant: 'primary' | 'secondary' | 'text', onClick: () => void): HTMLButtonElement => {
    const btn = document.createElement('button');
    btn.className = `dl-btn dl-btn-${variant}`;
    btn.innerText = text;
    btn.onclick = onClick;
    return btn;
};

export const renderToggle = (id: string, label: string, checked: boolean, disabled: boolean, onChange: (checked: boolean) => void): HTMLElement => {
    const wrapper = document.createElement('div');
    wrapper.className = 'dl-toggle-wrapper';

    const labelSpan = document.createElement('span');
    labelSpan.innerText = label;
    wrapper.appendChild(labelSpan);

    const switchLabel = document.createElement('label');
    switchLabel.className = 'dl-switch';

    const input = document.createElement('input');
    input.type = 'checkbox';
    input.id = id;
    input.checked = checked;
    input.disabled = disabled;
    input.onchange = (e) => onChange((e.target as HTMLInputElement).checked);

    const slider = document.createElement('span');
    slider.className = 'dl-slider';

    switchLabel.appendChild(input);
    switchLabel.appendChild(slider);
    wrapper.appendChild(switchLabel);

    return wrapper;
};
