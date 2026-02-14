import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@datalens/shared';
import { Input } from '@datalens/shared';
import { Textarea } from '@datalens/shared';
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@datalens/shared';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@datalens/shared';
import type { CreatePolicyRequest } from '../../types/governance';

const policySchema = z.object({
    name: z.string().min(1, 'Policy Name is required').max(100, 'Name too long'),
    type: z.enum(['retention', 'access', 'encryption', 'minimization']),
    description: z.string().optional(),
});

interface PolicyFormProps {
    onSubmit: (data: CreatePolicyRequest) => void;
    onCancel: () => void;
    isLoading?: boolean;
}

export const PolicyForm = ({ onSubmit, onCancel, isLoading }: PolicyFormProps) => {
    const form = useForm<z.infer<typeof policySchema>>({
        resolver: zodResolver(policySchema),
        defaultValues: {
            name: '',
            type: 'retention',
            description: '',
        },
    });

    const handleSubmit = (values: z.infer<typeof policySchema>) => {
        onSubmit({
            ...values,
            description: values.description || '',
            rules: {}, // Placeholder for now
        });
    };

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
                <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Policy Name</FormLabel>
                            <FormControl>
                                <Input placeholder="e.g., 7 Year Retention" {...field} />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="type"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Policy Type</FormLabel>
                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                <FormControl>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select a policy type" />
                                    </SelectTrigger>
                                </FormControl>
                                <SelectContent>
                                    <SelectItem value="retention">Retention</SelectItem>
                                    <SelectItem value="access">Access Control</SelectItem>
                                    <SelectItem value="encryption">Encryption</SelectItem>
                                    <SelectItem value="minimization">Data Minimization</SelectItem>
                                </SelectContent>
                            </Select>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="description"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Description</FormLabel>
                            <FormControl>
                                <Textarea
                                    placeholder="Describe the purpose of this policy..."
                                    className="resize-none"
                                    {...field}
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <div className="rounded-md bg-muted p-4 text-sm text-muted-foreground">
                    Configuration options for <strong>{form.watch('type')}</strong> will appear here.
                </div>

                <div className="flex justify-end gap-3 pt-4">
                    <Button type="button" variant="outline" onClick={onCancel} disabled={isLoading}>
                        Cancel
                    </Button>
                    <Button type="submit" disabled={isLoading}>
                        {isLoading ? 'Creating...' : 'Create Policy'}
                    </Button>
                </div>
            </form>
        </Form>
    );
};
