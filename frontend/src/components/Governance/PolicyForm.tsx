import React, { useState } from 'react';
import type { CreatePolicyRequest } from '../../types/governance';

interface PolicyFormProps {
    onSubmit: (data: CreatePolicyRequest) => void;
    onCancel: () => void;
    isLoading?: boolean;
}

export const PolicyForm: React.FC<PolicyFormProps> = ({ onSubmit, onCancel, isLoading }) => {
    const [formData, setFormData] = useState<CreatePolicyRequest>({
        name: '',
        type: 'retention',
        description: '',
        rules: {}
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <form onSubmit={handleSubmit} className="space-y-4">
            <div>
                <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                    Policy Name
                </label>
                <input
                    type="text"
                    id="name"
                    name="name"
                    required
                    value={formData.name}
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                    placeholder="e.g., 7 Year Retention"
                />
            </div>

            <div>
                <label htmlFor="type" className="block text-sm font-medium text-gray-700">
                    Policy Type
                </label>
                <select
                    id="type"
                    name="type"
                    value={formData.type}
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                >
                    <option value="retention">Retention</option>
                    <option value="access">Access Control</option>
                    <option value="encryption">Encryption</option>
                    <option value="minimization">Data Minimization</option>
                </select>
            </div>

            <div>
                <label htmlFor="description" className="block text-sm font-medium text-gray-700">
                    Description
                </label>
                <textarea
                    id="description"
                    name="description"
                    rows={3}
                    value={formData.description}
                    onChange={handleChange}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm p-2 border"
                    placeholder="Describe the purpose of this policy..."
                />
            </div>

            {/* Placeholder for dynamic rules based on type - keeping it simple for now */}
            <div>
                <label className="block text-sm font-medium text-gray-700">
                    Configuration
                </label>
                <div className="mt-1 p-3 bg-gray-50 rounded-md border border-gray-200 text-sm text-gray-500">
                    Configuration options for <strong>{formData.type}</strong> will appear here.
                </div>
            </div>

            <div className="flex justify-end space-x-3 pt-4 border-t border-gray-100">
                <button
                    type="button"
                    onClick={onCancel}
                    className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                    Cancel
                </button>
                <button
                    type="submit"
                    disabled={isLoading}
                    className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                >
                    {isLoading ? 'Creating...' : 'Create Policy'}
                </button>
            </div>
        </form>
    );
};
