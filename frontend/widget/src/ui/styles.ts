import { ThemeConfig } from '../types.ts';

export function injectStyles(theme: ThemeConfig, customCss?: string) {
    const styleId = 'datalens-consent-styles';
    if (document.getElementById(styleId)) return;

    const css = `
        :root {
            --dl-primary: ${theme.primary_color};
            --dl-bg: ${theme.background_color};
            --dl-text: ${theme.text_color};
            --dl-font: ${theme.font_family};
            --dl-radius: ${theme.border_radius};
        }

        .dl-consent-container {
            font-family: var(--dl-font);
            background-color: var(--dl-bg);
            color: var(--dl-text);
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
            padding: 1rem;
            position: fixed;
            z-index: 9999;
            box-sizing: border-box;
        }

        .dl-consent-container * {
            box-sizing: border-box;
        }

        .dl-consent-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.5rem;
        }
        
        .dl-consent-title {
            font-weight: 600;
            font-size: 1.125rem;
            margin: 0;
        }

        .dl-consent-description {
            font-size: 0.875rem;
            margin-bottom: 1rem;
            line-height: 1.5;
        }

        .dl-consent-actions {
            display: flex;
            gap: 0.5rem;
            justify-content: flex-end;
        }

        .dl-btn {
            padding: 0.5rem 1rem;
            border-radius: var(--dl-radius);
            font-size: 0.875rem;
            font-weight: 500;
            cursor: pointer;
            border: 1px solid transparent;
            transition: opacity 0.2s;
        }

        .dl-btn:hover {
            opacity: 0.9;
        }

        .dl-btn-primary {
            background-color: var(--dl-primary);
            color: white; /* Basic contrast, improvement: calculate based on bg */
        }

        .dl-btn-secondary {
            background-color: transparent;
            border-color: var(--dl-text);
            color: var(--dl-text);
        }

        .dl-btn-text {
            background-color: transparent;
            color: var(--dl-text);
            text-decoration: underline;
        }

        /* Layout Specifics */
        .dl-layout-bottom-bar {
            bottom: 0;
            left: 0;
            right: 0;
            border-top: 1px solid rgba(0,0,0,0.1);
        }

        .dl-layout-top-bar {
            top: 0;
            left: 0;
            right: 0;
            border-bottom: 1px solid rgba(0,0,0,0.1);
        }

        .dl-layout-modal {
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            max-width: 500px;
            width: 90%;
            border-radius: var(--dl-radius);
            box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
        }
        
        .dl-backdrop {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0,0,0,0.5);
            z-index: 9998;
        }

        /* Toggles */
        .dl-toggle-wrapper {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 0.5rem;
            padding: 0.5rem;
            background: rgba(0,0,0,0.03);
            border-radius: 4px;
        }

        .dl-switch {
            position: relative;
            display: inline-block;
            width: 36px;
            height: 20px;
        }

        .dl-switch input { 
            opacity: 0;
            width: 0;
            height: 0;
        }

        .dl-slider {
            position: absolute;
            cursor: pointer;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: #ccc;
            transition: .4s;
            border-radius: 20px;
        }

        .dl-slider:before {
            position: absolute;
            content: "";
            height: 16px;
            width: 16px;
            left: 2px;
            bottom: 2px;
            background-color: white;
            transition: .4s;
            border-radius: 50%;
        }

        input:checked + .dl-slider {
            background-color: var(--dl-primary);
        }

        input:focus + .dl-slider {
            box-shadow: 0 0 1px var(--dl-primary);
        }

        input:checked + .dl-slider:before {
            transform: translateX(16px);
        }

        ${customCss || ''}
    `;

    const style = document.createElement('style');
    style.id = styleId;
    style.textContent = css;
    document.head.appendChild(style);
}
