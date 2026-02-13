import { useState } from 'react';
import { adminService } from '../../../services/adminService';
import { toast } from '../../../stores/toastStore';
import { Button } from '../../../components/common/Button';
import { Modal } from '../../../components/common/Modal';
import type { CreateTenantInput } from '../../../types/admin';

interface TenantFormProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
}

export function TenantForm({ isOpen, onClose, onSuccess }: TenantFormProps) {
    const [formData, setFormData] = useState<CreateTenantInput>({
        name: '',
        domain: '',
        admin_email: '',
        plan: 'FREE',
    });
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSubmitting(true);
        try {
            await adminService.createTenant(formData);
            toast.success('Success', `Tenant "${formData.name}" created successfully.`);
            onSuccess();
            onClose();
            // Reset form
            setFormData({ name: '', domain: '', admin_email: '', plan: 'FREE' });
        } catch (error) {
            console.error('Failed to create tenant', error);
            // Error handling is done globally in api interceptor for toasts,
            // but we can add specific form errors here if needed
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <Modal
            open={isOpen}
            onClose={onClose}
            title="Onboard New Tenant"
            size="lg"
        >
            <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Organization Name
                    </label>
                    <input
                        type="text"
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        placeholder="Acme Corp"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Subdomain
                    </label>
                    <div className="flex">
                        <input
                            type="text"
                            required
                            pattern="[a-z0-9-]+"
                            title="Lowercase letters, numbers, and hyphens only"
                            className="flex-1 px-3 py-2 border border-gray-300 rounded-l-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                            value={formData.domain}
                            onChange={(e) => setFormData({ ...formData, domain: e.target.value.toLowerCase() })}
                            placeholder="acme"
                        />
                        <span className="inline-flex items-center px-3 rounded-r-md border border-l-0 border-gray-300 bg-gray-50 text-gray-500 text-sm">
                            .datalens.com
                        </span>
                    </div>
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Admin Email
                    </label>
                    <input
                        type="email"
                        required
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
                        value={formData.admin_email}
                        onChange={(e) => setFormData({ ...formData, admin_email: e.target.value })}
                        placeholder="admin@acme.com"
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Subscription Plan
                    </label>
                    <select
                        className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 bg-white"
                        value={formData.plan}
                        onChange={(e) => setFormData({ ...formData, plan: e.target.value as CreateTenantInput['plan'] })}
                    >
                        <option value="FREE">Free</option>
                        <option value="STARTER">Starter</option>
                        <option value="PROFESSIONAL">Professional</option>
                        <option value="ENTERPRISE">Enterprise</option>
                    </select>
                </div>

                <div className="pt-4 flex justify-end space-x-3">
                    <Button variant="secondary" onClick={onClose} type="button">
                        Cancel
                    </Button>
                    <Button type="submit" isLoading={isSubmitting}>
                        Create Tenant
                    </Button>
                </div>
            </form>
        </Modal>
    );
}
