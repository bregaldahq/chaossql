export async function submitLead(payload: Record<string, unknown>): Promise<void> {
  const response = await fetch('/api/waitlist', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
  });
  if (!response.ok) throw new Error(`HTTP ${response.status}: Request delivery failed. Please try again.`);
  const result = await response.json();
  if (result.success !== true || result.lead?.dispatched !== true) {
    throw new Error('Request delivery was not acknowledged. Please try again.');
  }
}
