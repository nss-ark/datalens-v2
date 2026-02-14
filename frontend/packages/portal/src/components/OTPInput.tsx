import React, { useRef, useEffect } from 'react';

interface OTPInputProps {
    length?: number;
    value: string;
    onChange: (value: string) => void;
}

export const OTPInput: React.FC<OTPInputProps> = ({ length = 6, value, onChange }) => {
    const inputs = useRef<(HTMLInputElement | null)[]>([]);

    useEffect(() => {
        if (inputs.current[0]) {
            inputs.current[0].focus();
        }
    }, []);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>, index: number) => {
        const val = e.target.value;
        if (isNaN(Number(val))) return;

        const newOtp = value.split('');
        newOtp[index] = val.substring(val.length - 1);
        const newValue = newOtp.join('');
        onChange(newValue);

        // Focus next input
        if (val && index < length - 1 && inputs.current[index + 1]) {
            inputs.current[index + 1]?.focus();
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>, index: number) => {
        if (e.key === 'Backspace' && !value[index] && index > 0 && inputs.current[index - 1]) {
            inputs.current[index - 1]?.focus();
        }
    };

    // Ensure value matches length
    const otpArray = Array(length).fill('');
    for (let i = 0; i < length; i++) {
        otpArray[i] = value[i] || '';
    }

    return (
        <div className="flex gap-2 justify-center">
            {otpArray.map((digit, index) => (
                <input
                    key={index}
                    ref={(el) => { inputs.current[index] = el; }}
                    type="text"
                    maxLength={1}
                    value={digit}
                    onChange={(e) => handleChange(e, index)}
                    onKeyDown={(e) => handleKeyDown(e, index)}
                    className="w-12 h-12 text-center text-xl font-semibold border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition-all"
                />
            ))}
        </div>
    );
};
