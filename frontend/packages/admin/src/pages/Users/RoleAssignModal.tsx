import { useState, useEffect } from 'react';
import { adminService } from '@/services/adminService';
import { toast } from '@datalens/shared';
import { Button } from '@datalens/shared';
import { Modal } from '@datalens/shared';
import type { AdminUser, AdminRole } from '@/types/admin';

interface RoleAssignModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSuccess: () => void;
    user: AdminUser | null;
}

export function RoleAssignModal({ isOpen, onClose, onSuccess, user }: RoleAssignModalProps) {
    const [roles, setRoles] = useState<AdminRole[]>([]);
    const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isLoadingRoles, setIsLoadingRoles] = useState(false);

    useEffect(() => {
        if (isOpen) {
            loadRoles();
            if (user) {
                setSelectedRoleIds(user.role_ids || []);
            }
        }
    }, [isOpen, user]);

    const loadRoles = async () => {
        setIsLoadingRoles(true);
        try {
            const data = await adminService.getRoles();
            setRoles(data);
        } catch (error) {
            console.error('Failed to load roles', error);
            // Toast handled by interceptor usually, but good to have fallback
        } finally {
            setIsLoadingRoles(false);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!user) return;

        setIsSubmitting(true);
        try {
            await adminService.assignRoles(user.id, selectedRoleIds);
            toast.success('Success', `Roles updated for ${user.name}`);
            onSuccess();
            onClose();
        } catch (error) {
            console.error('Failed to assign roles', error);
        } finally {
            setIsSubmitting(false);
        }
    };

    const toggleRole = (roleId: string) => {
        setSelectedRoleIds(prev =>
            prev.includes(roleId)
                ? prev.filter(id => id !== roleId)
                : [...prev, roleId]
        );
    };

    return (
        <Modal
            open={isOpen}
            onClose={onClose}
            title={`Manage Roles: ${user?.name || ''}`}
            size="md"
        >
            <form onSubmit={handleSubmit} className="space-y-4">
                <div className="space-y-2">
                    {isLoadingRoles ? (
                        <div className="text-gray-500 text-sm">Loading roles...</div>
                    ) : (
                        roles.map(role => (
                            <div key={role.id} className="flex items-start space-x-3 p-2 hover:bg-gray-50 rounded">
                                <input
                                    type="checkbox"
                                    id={`role-${role.id}`}
                                    checked={selectedRoleIds.includes(role.id)}
                                    onChange={() => toggleRole(role.id)}
                                    className="mt-1 h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                                />
                                <label htmlFor={`role-${role.id}`} className="flex-1 cursor-pointer">
                                    <div className="font-medium text-gray-900">{role.name}</div>
                                    <div className="text-xs text-gray-500">{role.description}</div>
                                </label>
                            </div>
                        ))
                    )}
                </div>

                <div className="pt-4 flex justify-end space-x-3 border-t border-gray-100 mt-4">
                    <Button variant="secondary" onClick={onClose} type="button">
                        Cancel
                    </Button>
                    <Button type="submit" isLoading={isSubmitting}>
                        Save Changes
                    </Button>
                </div>
            </form>
        </Modal>
    );
}
