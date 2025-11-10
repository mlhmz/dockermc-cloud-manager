import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { useCreateServer } from '@/hooks/mutations/useCreateServer';

const createServerSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .max(100, 'Name must be less than 100 characters')
    .regex(/^[a-zA-Z0-9-_]+$/, 'Name can only contain letters, numbers, hyphens, and underscores'),
  max_players: z
    .number()
    .int()
    .min(1, 'Minimum 1 player')
    .max(1000, 'Maximum 1000 players')
    .optional(),
  motd: z
    .string()
    .max(255, 'MOTD must be less than 255 characters')
    .optional(),
  version: z
    .string()
    .optional(),
});

type CreateServerFormValues = z.infer<typeof createServerSchema>;

interface CreateServerDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateServerDialog({ open, onOpenChange }: CreateServerDialogProps) {
  const createServer = useCreateServer();

  const form = useForm<CreateServerFormValues>({
    resolver: zodResolver(createServerSchema),
    defaultValues: {
      name: '',
      max_players: 20,
      motd: '',
      version: 'LATEST',
    },
  });

  const onSubmit = async (data: CreateServerFormValues) => {
    try {
      await createServer.mutateAsync(data);
      form.reset();
      onOpenChange(false);
    } catch (error) {
      // Error is handled by the mutation
      console.error('Failed to create server:', error);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Create New Server</DialogTitle>
          <DialogDescription>
            Create a new Minecraft server. The server will be automatically connected to the proxy.
          </DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Server Name</FormLabel>
                  <FormControl>
                    <Input placeholder="survival-server" {...field} />
                  </FormControl>
                  <FormDescription>
                    A unique name for your server (letters, numbers, hyphens, underscores only)
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="max_players"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Max Players</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      {...field}
                      onChange={(e) => field.onChange(parseInt(e.target.value) || 0)}
                    />
                  </FormControl>
                  <FormDescription>
                    Maximum number of players (1-1000)
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="motd"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Message of the Day (MOTD)</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Welcome to my Minecraft server!"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    The message displayed in the server list (optional)
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="version"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Minecraft Version</FormLabel>
                  <FormControl>
                    <Input placeholder="LATEST" {...field} />
                  </FormControl>
                  <FormDescription>
                    Minecraft version (e.g., "1.20.1" or "LATEST")
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={createServer.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={createServer.isPending}>
                {createServer.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                Create Server
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
