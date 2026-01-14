import toast from 'react-hot-toast';
export { generateViewShareUrl, generateDomainShareUrl } from '../lib/deepLinks';

export async function copyToClipboard(text: string, successMessage = 'Link copied to clipboard'): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    toast.success(successMessage);
    return true;
  } catch {
    toast.error('Failed to copy to clipboard');
    return false;
  }
}
