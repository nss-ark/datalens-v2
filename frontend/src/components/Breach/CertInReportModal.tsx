import { Modal } from '../common/Modal';
import { Button } from '../common/Button';
import { Copy } from 'lucide-react';
import { toast } from 'react-toastify';

interface CertInReportModalProps {
    isOpen: boolean;
    onClose: () => void;
    reportData: Record<string, unknown> | null;
}

export const CertInReportModal = ({ isOpen, onClose, reportData }: CertInReportModalProps) => {
    const handleCopy = () => {
        if (reportData) {
            navigator.clipboard.writeText(JSON.stringify(reportData, null, 2));
            toast.success('Report JSON copied to clipboard');
        }
    };

    return (
        <Modal open={isOpen} onClose={onClose} title="CERT-In Incident Report">
            <div className="space-y-4">
                <div className="bg-gray-50 p-4 rounded-md border text-sm font-mono overflow-auto max-h-[400px]">
                    <pre>{reportData ? JSON.stringify(reportData, null, 2) : 'Loading...'}</pre>
                </div>
                <div className="flex justify-end gap-2">
                    <Button variant="outline" onClick={onClose}>Close</Button>
                    <Button icon={<Copy size={16} />} onClick={handleCopy}>
                        Copy JSON
                    </Button>
                </div>
            </div>
        </Modal>
    );
};
