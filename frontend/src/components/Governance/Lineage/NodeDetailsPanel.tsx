import React from 'react';
import { X, Database, Shield, Activity } from 'lucide-react';
import type { GraphNode } from '../../../types/lineage';
import { StatusBadge } from '../../common/StatusBadge';
import { Button } from '../../common/Button';

interface NodeDetailsPanelProps {
    node: GraphNode | null;
    onClose: () => void;
}

const NodeDetailsPanel: React.FC<NodeDetailsPanelProps> = ({ node, onClose }) => {
    if (!node) return null;

    const { data } = node;

    return (
        <div className="fixed inset-y-0 right-0 w-96 bg-white shadow-2xl border-l border-gray-200 transform transition-transform duration-300 ease-in-out z-50 overflow-y-auto">
            <div className="p-6">
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-bold text-gray-900">Node Details</h2>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
                        <X size={24} />
                    </button>
                </div>

                <div className="space-y-6">
                    {/* Header Info */}
                    <div className="flex items-start space-x-4">
                        <div className="p-3 bg-blue-50 rounded-lg">
                            <Database className="w-6 h-6 text-blue-600" />
                        </div>
                        <div>
                            <h3 className="text-lg font-medium text-gray-900">{node.label}</h3>
                            <p className="text-sm text-gray-500 uppercase tracking-wide">{node.type}</p>
                        </div>
                    </div>

                    {/* Metadata Grid */}
                    <div className="grid grid-cols-1 gap-4 bg-gray-50 p-4 rounded-lg border border-gray-100">
                        <div>
                            <span className="text-xs text-gray-500 uppercase font-semibold">Risk Level</span>
                            <div className="mt-1">
                                <StatusBadge label={(data?.riskLevel as string) || 'LOW'} />
                            </div>
                        </div>
                        <div>
                            <span className="text-xs text-gray-500 uppercase font-semibold">Owner</span>
                            <p className="text-sm font-medium text-gray-900 mt-1">{(data?.owner as string) || 'Unknown'}</p>
                        </div>
                        <div>
                            <span className="text-xs text-gray-500 uppercase font-semibold">Last Verified</span>
                            <p className="text-sm font-medium text-gray-900 mt-1">{(data?.lastVerified as string) || 'Never'}</p>
                        </div>
                    </div>

                    {/* Sensitivity Info */}
                    <div>
                        <h4 className="flex items-center text-sm font-semibold text-gray-900 mb-3">
                            <Shield className="w-4 h-4 mr-2 text-gray-500" />
                            Data Sensitivity
                        </h4>
                        <div className="flex flex-wrap gap-2">
                            {(data?.piiTypes as string[])?.map((tag: string) => (
                                <span key={tag} className="px-2 py-1 bg-purple-100 text-purple-700 text-xs rounded-full border border-purple-200">
                                    {tag}
                                </span>
                            )) || <span className="text-sm text-gray-500">No PII detected</span>}
                        </div>
                    </div>

                    {/* Lineage Info */}
                    <div>
                        <h4 className="flex items-center text-sm font-semibold text-gray-900 mb-3">
                            <Activity className="w-4 h-4 mr-2 text-gray-500" />
                            Processing Purpose
                        </h4>
                        <p className="text-sm text-gray-600 leading-relaxed">
                            {(data?.purpose as string) || 'No extensive processing documented for this node.'}
                        </p>
                    </div>

                    <div className="pt-4 border-t border-gray-200">
                        <Button variant="secondary" className="w-full">View Full Asset Record</Button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default NodeDetailsPanel;
