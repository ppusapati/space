import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import Input from '../Input.svelte';

describe('Input', () => {
  it('renders with default props', () => {
    render(Input);
    const input = screen.getByRole('textbox');
    expect(input).toBeInTheDocument();
  });

  it('renders with label', () => {
    render(Input, { props: { label: 'Email', id: 'email-input' } });
    const label = screen.getByText('Email');
    expect(label).toBeInTheDocument();
    expect(label).toHaveAttribute('for', 'email-input');
  });

  it('renders required indicator when required', () => {
    render(Input, { props: { label: 'Name', required: true } });
    expect(screen.getByText('*')).toBeInTheDocument();
  });

  it('renders with placeholder', () => {
    render(Input, { props: { placeholder: 'Enter your name' } });
    expect(screen.getByPlaceholderText('Enter your name')).toBeInTheDocument();
  });

  it('renders with initial value', () => {
    render(Input, { props: { value: 'Hello World' } });
    expect(screen.getByDisplayValue('Hello World')).toBeInTheDocument();
  });

  it('emits input event on typing', async () => {
    const handleInput = vi.fn();
    const { component } = render(Input);
    component.$on('input', handleInput);

    const input = screen.getByRole('textbox');
    await fireEvent.input(input, { target: { value: 'test' } });

    expect(handleInput).toHaveBeenCalled();
  });

  it('emits change event on change', async () => {
    const handleChange = vi.fn();
    const { component } = render(Input);
    component.$on('change', handleChange);

    const input = screen.getByRole('textbox');
    await fireEvent.change(input, { target: { value: 'test' } });

    expect(handleChange).toHaveBeenCalled();
  });

  it('emits focus and blur events', async () => {
    const handleFocus = vi.fn();
    const handleBlur = vi.fn();
    const { component } = render(Input);
    component.$on('focus', handleFocus);
    component.$on('blur', handleBlur);

    const input = screen.getByRole('textbox');
    await fireEvent.focus(input);
    expect(handleFocus).toHaveBeenCalled();

    await fireEvent.blur(input);
    expect(handleBlur).toHaveBeenCalled();
  });

  it('is disabled when disabled prop is true', () => {
    render(Input, { props: { disabled: true } });
    expect(screen.getByRole('textbox')).toBeDisabled();
  });

  it('is readonly when readonly prop is true', () => {
    render(Input, { props: { readonly: true } });
    expect(screen.getByRole('textbox')).toHaveAttribute('readonly');
  });

  it('shows helper text', () => {
    render(Input, { props: { helperText: 'Enter a valid email', id: 'test-input' } });
    expect(screen.getByText('Enter a valid email')).toBeInTheDocument();
  });

  it('shows error text when state is invalid', () => {
    render(Input, {
      props: {
        state: 'invalid',
        errorText: 'This field is required',
        id: 'error-input',
      },
    });
    expect(screen.getByText('This field is required')).toBeInTheDocument();
  });

  it('has aria-invalid when state is invalid', () => {
    render(Input, { props: { state: 'invalid' } });
    expect(screen.getByRole('textbox')).toHaveAttribute('aria-invalid', 'true');
  });

  it('shows clear button when clearable and has value', async () => {
    render(Input, { props: { clearable: true, value: 'some text' } });
    const clearButton = screen.getByLabelText('Clear input');
    expect(clearButton).toBeInTheDocument();
  });

  it('clears value when clear button is clicked', async () => {
    const handleClear = vi.fn();
    const { component } = render(Input, { props: { clearable: true, value: 'some text' } });
    component.$on('clear', handleClear);

    const clearButton = screen.getByLabelText('Clear input');
    await fireEvent.click(clearButton);

    expect(handleClear).toHaveBeenCalled();
  });

  it('does not show clear button when disabled', () => {
    render(Input, { props: { clearable: true, value: 'some text', disabled: true } });
    expect(screen.queryByLabelText('Clear input')).not.toBeInTheDocument();
  });

  it('renders number input type correctly', () => {
    render(Input, { props: { type: 'number', min: 0, max: 100, step: 1 } });
    const input = screen.getByRole('spinbutton');
    expect(input).toHaveAttribute('type', 'number');
    expect(input).toHaveAttribute('min', '0');
    expect(input).toHaveAttribute('max', '100');
  });

  it('applies custom class', () => {
    render(Input, { props: { class: 'custom-class' } });
    const input = screen.getByRole('textbox');
    expect(input).toHaveClass('custom-class');
  });

  it('renders with test-id', () => {
    render(Input, { props: { testId: 'my-input' } });
    expect(screen.getByTestId('my-input')).toBeInTheDocument();
  });
});
